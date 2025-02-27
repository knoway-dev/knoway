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

func (rl *RateLimiter) log(level slog.Level, msg string, args ...any) {
	commonArgs := []any{
		"filter", "rate_limit",
		"serverPrefix", rl.serverPrefix,
		"mode", rl.mode,
	}
	args = append(commonArgs, args...)
	slog.Log(context.Background(), level, msg, args...)
}

func (rl *RateLimiter) Info(msg string, args ...any) {
	rl.log(slog.LevelInfo, msg, args...)
}

func (rl *RateLimiter) Debug(msg string, args ...any) {
	rl.log(slog.LevelDebug, msg, args...)
}

func (rl *RateLimiter) Error(msg string, args ...any) {
	rl.log(slog.LevelError, msg, args...)
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

	rl.Info("initializing rate limiter")
	rl.Debug("rate limiter default policies",
		"serverPrefix", rl.serverPrefix,
		"mode", rl.mode,
		"pluginPolicies", rl.pluginPolicies)

	if rl.mode == v1alpha1.RateLimitMode_REDIS {
		rl.Info("initializing redis client", "url", rCfg.GetRedisServer().GetUrl())

		redisClient, err := redis.NewRedisClient(rCfg.GetRedisServer().GetUrl())
		if err != nil {
			rl.Error("failed to create redis client", "error", err)
			return nil, fmt.Errorf("failed to create redis client: %w", err)
		}

		rl.redisClient = redisClient
	} else {
		rl.Info("initializing local rate limiter shards")
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
			rl.Info("stopping rate limiter")
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
		rl.Debug("no api key or user name found, skipping rate limit")
		return filters.NewOK()
	}

	route := rMeta.MatchRoute
	routeName := route.GetName()

	var fPolicy *v1alpha1.RateLimitPolicy
	if routeName == "" {
		routeName = request.GetModel()
	}

	var rCfg *v1alpha1.RateLimitConfig

	for _, f := range route.GetFilters() {
		newRl, _ := NewRateLimitConfigWithFilter(f.GetConfig())
		if newRl != nil {
			rCfg = newRl
			break
		}
	}

	if rCfg != nil {
		fPolicy = rl.findMatchingPolicy(apiKey, userName, rCfg.GetPolicies())
	}
	if fPolicy == nil {
		fPolicy = rl.findMatchingPolicy(apiKey, userName, rl.pluginPolicies)
	}

	allow, err := rl.allowRequest(apiKey, userName, routeName, fPolicy)
	if err != nil {
		rl.Error("failed to check rate limit", "error", err)
		return filters.NewFailed(err)
	}

	if !allow {
		rl.Debug("rate limit exceeded",
			"apiKey", apiKey,
			"userName", userName,
			"route", routeName,
			"limit", fPolicy.GetLimit(),
			"duration", fPolicy.GetDuration().AsDuration())

		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
}

func (rl *RateLimiter) allowRequest(apiKey, userName string, routeName string, policy *v1alpha1.RateLimitPolicy) (bool, error) {
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

	key := rl.buildKey(policy.GetBasedOn(), value, routeName)

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
