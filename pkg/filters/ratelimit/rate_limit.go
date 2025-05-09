package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/redis"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/protoutils"

	"github.com/redis/rueidis"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	cleanupInterval = 30 * time.Minute
	maxTTL          = 5 * time.Minute
	ttlRate         = 2

	numShards          = 64    // Number of shards, must be power of 2
	maxBucketsPerShard = 10000 // Maximum buckets per shard

	precision           = 1000 // Precision for fixed-point arithmetic
	defaultDuration     = 1 * time.Minute
	defaultServerPrefix = "knoway-rate-limit"
)

type RateLimiter struct {
	filters.IsRequestFilter

	shards    []*rateLimitShard
	numShards int
	cancel    context.CancelFunc

	pluginPolicies []*v1alpha1.RateLimitPolicy
	mode           v1alpha1.RateLimitMode

	serverPrefix string

	redisClient rueidis.Client
}

func (rl *RateLimiter) logCommonAttrs() []any {
	return []any{
		slog.String("filter", "rate_limit"),
		slog.String("serverPrefix", rl.serverPrefix),
		slog.Any("mode", rl.mode),
	}
}

var _ filters.RequestFilter = (*RateLimiter)(nil)
var _ filters.OnCompletionRequestFilter = (*RateLimiter)(nil)
var _ filters.OnImageGenerationsRequestFilter = (*RateLimiter)(nil)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	rCfg, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		slog.Error("invalid rate limit config", "error", err)
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		shards:       make([]*rateLimitShard, numShards),
		serverPrefix: rCfg.GetServerPrefix(),
		numShards:    numShards,
		cancel:       cancel,

		pluginPolicies: rCfg.GetPolicies(),
		mode:           rCfg.GetModel(),
	}

	if rl.serverPrefix == "" {
		rl.serverPrefix = defaultServerPrefix
	}

	if rl.mode == v1alpha1.RateLimitMode_RATE_LIMIT_MODEL_UNSPECIFIED {
		rl.mode = v1alpha1.RateLimitMode_LOCAL
	}

	slog.InfoContext(context.Background(), "initializing rate limiter", rl.logCommonAttrs()...)
	slog.DebugContext(context.Background(), "rate limiter default policies", append(rl.logCommonAttrs(), slog.Any("pluginPolicies", rl.pluginPolicies))...)

	if rl.mode == v1alpha1.RateLimitMode_REDIS {
		slog.InfoContext(context.Background(), "initializing redis client", append(rl.logCommonAttrs(), slog.String("url", rCfg.GetRedisServer().GetUrl()))...)

		redisClient, err := redis.NewRedisClient(rCfg.GetRedisServer().GetUrl())
		if err != nil {
			slog.ErrorContext(context.Background(), "failed to create redis client", append(rl.logCommonAttrs(), slog.Any("error", err))...)
			return nil, fmt.Errorf("failed to create redis client: %w", err)
		}

		rl.redisClient = redisClient
	} else {
		slog.InfoContext(context.Background(), "initializing local rate limiter shards", rl.logCommonAttrs()...)
		// init shards for local mode
		for i := range numShards {
			rl.shards[i] = &rateLimitShard{
				buckets:        make(map[string]*tokenBucket),
				lastAccessTime: make(map[string]time.Time),
			}
		}

		// start cleanup
		go rl.cleanupLoop(ctx)
	}

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStop: func(ctx context.Context) error {
			slog.InfoContext(context.Background(), "stopping rate limiter", rl.logCommonAttrs()...)
			rl.cancel()
			if rl.redisClient != nil {
				rl.redisClient.Close()
			}
			return nil
		},
	})

	return rl, nil
}

func (rl *RateLimiter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	return rl.onRequest(ctx, request)
}

func (rl *RateLimiter) OnImageGenerationsRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	return rl.onRequest(ctx, request)
}

func (rl *RateLimiter) buildKey(baseOn v1alpha1.RateLimitBaseOn, value string, routeName string) string {
	return fmt.Sprintf("%s:%s:%s:%s", rl.serverPrefix, baseOn, value, routeName)
}

func NewRateLimitConfigWithFilter(cfg *anypb.Any) (*v1alpha1.RateLimitConfig, error) {
	if cfg == nil {
		return nil, nil
	}

	res, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return res, nil
}

func (rl *RateLimiter) findMatchingPolicy(apiKey, userName string, policies []*v1alpha1.RateLimitPolicy) *v1alpha1.RateLimitPolicy {
	if policies == nil {
		return nil
	}

	for i, policy := range policies {
		var value string

		switch policy.GetBasedOn() {
		case v1alpha1.RateLimitBaseOn_API_KEY:
			value = apiKey
		case v1alpha1.RateLimitBaseOn_USER_ID:
			value = userName
		case v1alpha1.RateLimitBaseOn_RATE_LIMIT_BASE_ON_UNSPECIFIED:
			continue
		default:
			continue
		}

		matched := false
		if policy.GetMatch() == nil {
			// effective scope: any baseOn value
			matched = true
		} else {
			if policy.GetMatch().GetExact() == value {
				matched = true
			} else if policy.GetMatch().GetPrefix() != "" && strings.HasPrefix(value, policy.GetMatch().GetPrefix()) {
				matched = true
			}
		}

		if matched {
			return policies[i]
		}
	}

	return nil
}

func (rl *RateLimiter) onRequest(ctx context.Context, request object.LLMRequest) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	apiKey := rMeta.AuthInfo.GetApiKeyId()
	userName := rMeta.AuthInfo.GetUserId()

	if apiKey == "" && userName == "" {
		slog.DebugContext(context.Background(), "no api key or user name found, skipping rate limit", rl.logCommonAttrs()...)
		return filters.NewOK()
	}

	fPolicy := rl.findMatchingPolicy(apiKey, userName, rl.pluginPolicies)
	if fPolicy == nil {
		slog.DebugContext(ctx, "no matching policy found, skipping rate limit", append(rl.logCommonAttrs(), slog.String("apiKey", apiKey), slog.String("userName", userName))...)
		return filters.NewOK()
	}

	allow, err := rl.allowRequest(apiKey, userName, request.GetModel(), fPolicy)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check rate limit", append(rl.logCommonAttrs(), slog.Any("error", err))...)
		return filters.NewFailed(err)
	}

	if !allow {
		slog.DebugContext(ctx, "rate limit exceeded", append(
			rl.logCommonAttrs(),
			slog.String("apiKey", apiKey),
			slog.String("userName", userName),
			slog.String("model", request.GetModel()),
			slog.Int64("limit", int64(fPolicy.GetLimit())),
			slog.Duration("duration", fPolicy.GetDuration().AsDuration()),
		)...)

		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
}

func (rl *RateLimiter) allowRequest(apiKey, userName string, modelName string, policy *v1alpha1.RateLimitPolicy) (bool, error) {
	if policy == nil {
		return true, nil
	}

	var value string

	switch policy.GetBasedOn() {
	case v1alpha1.RateLimitBaseOn_API_KEY:
		value = apiKey
	case v1alpha1.RateLimitBaseOn_USER_ID:
		value = userName
	case v1alpha1.RateLimitBaseOn_RATE_LIMIT_BASE_ON_UNSPECIFIED:
		return true, nil
	default:
		return true, nil
	}

	matched := false
	if policy.GetMatch() == nil {
		// effective scope: any baseOn value
		matched = true
	} else {
		if policy.GetMatch().GetExact() == value {
			matched = true
		} else if policy.GetMatch().GetPrefix() != "" && strings.HasPrefix(value, policy.GetMatch().GetPrefix()) {
			matched = true
		}
	}

	if !matched {
		return true, nil
	}

	// disabled limit
	if policy.GetLimit() == 0 {
		return true, nil
	}

	duration := policy.GetDuration().AsDuration()
	if duration == 0 {
		duration = defaultDuration
	}

	key := rl.buildKey(policy.GetBasedOn(), value, modelName)

	return rl.checkBucket(key, duration, int(policy.GetLimit()))
}

func (rl *RateLimiter) checkBucket(key string, window time.Duration, limit int) (bool, error) {
	if limit == 0 {
		return true, nil
	}

	if window.Seconds() == 0 {
		window = defaultDuration
	}

	if rl.mode == v1alpha1.RateLimitMode_REDIS {
		return rl.checkBucketRedis(key, window, limit)
	}

	return rl.checkBucketLocal(key, window, limit)
}
