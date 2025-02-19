package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"knoway.dev/pkg/object"

	"google.golang.org/protobuf/types/known/anypb"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/protoutils"

	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
)

// todo GetRoutePolicies From CRD
var modelRouteRateLimitConfigs map[string]RateLimitConfig

func getModelRouteRateLimitConfigs(modelName string) (RateLimitConfig, bool) {
	if modelRouteRateLimitConfigs == nil {
		return RateLimitConfig{}, false
	}
	if cfg, ok := modelRouteRateLimitConfigs[modelName]; ok {
		return cfg, true
	}

	return RateLimitConfig{}, false
}

const (
	defaultRateLimit       = 50
	defaultWindowInSeconds = 10
)

var defaultPolicy = &v1alpha1.RateLimitConfig_Policy{
	BaseOn:   v1alpha1.RateLimitConfig_API_KEY,
	Count:    defaultRateLimit,
	Internal: defaultWindowInSeconds, // 60 seconds
}

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	c, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	if c.GetDefaultPolicy() == nil {
		c.DefaultPolicy = defaultPolicy
	}

	config := RateLimitConfig{
		strategy: c.GetDefaultPolicy().GetBaseOn(),
		count:    int(c.GetDefaultPolicy().GetCount()),
		window:   time.Duration(c.GetDefaultPolicy().GetInternal()) * time.Second,
	}

	return &RateLimiter{
		globalConfig: config,
		apiKeyBucket: make(map[string][]time.Time),
		userBucket:   make(map[string][]time.Time),
	}, nil
}

type RateLimitConfig struct {
	strategy v1alpha1.RateLimitConfig_Strategy
	count    int
	window   time.Duration
}

var _ filters.RequestFilter = (*RateLimiter)(nil)
var _ filters.OnCompletionRequestFilter = (*RateLimiter)(nil)

type RateLimiter struct {
	filters.IsRequestFilter

	globalConfig RateLimitConfig
	apiKeyBucket map[string][]time.Time
	userBucket   map[string][]time.Time
	mu           sync.Mutex
}

func (rl *RateLimiter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	apiKey := rMeta.AuthInfo.GetApiKeyId()
	userName := rMeta.AuthInfo.GetUserId()

	modelName := request.GetModel()
	mCfg, find := getModelRouteRateLimitConfigs(modelName)

	if !find {
		mCfg = rl.globalConfig
	}

	if !rl.AllowRequestWithConfig(apiKey, userName, mCfg) {
		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
}

func (rl *RateLimiter) checkBucket(bucket map[string][]time.Time, key string, window time.Duration, count int) bool {
	if window == 0 {
		window = time.Second
	}
	now := time.Now()
	windowStart := now.Add(-window)
	records := bucket[key]
	var validRecords []time.Time

	for _, record := range records {
		if record.After(windowStart) {
			validRecords = append(validRecords, record)
		}
	}
	bucket[key] = validRecords

	if len(validRecords) >= count {
		return false
	}

	bucket[key] = append(bucket[key], now)

	return true
}

func (rl *RateLimiter) AllowRequestWithConfig(apiKey, userName string, config RateLimitConfig) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	switch config.strategy {
	case v1alpha1.RateLimitConfig_API_KEY:
		return rl.checkBucket(rl.apiKeyBucket, apiKey, config.window, config.count)
	case v1alpha1.RateLimitConfig_USER:
		return rl.checkBucket(rl.userBucket, userName, config.window, config.count)
	case v1alpha1.RateLimitConfig_API_KEY_AND_USER:
		if !rl.checkBucket(rl.apiKeyBucket, apiKey, config.window, config.count) {
			return false
		}

		return rl.checkBucket(rl.userBucket, userName, config.window, config.count)
	default:
		return true
	}
}
