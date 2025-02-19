package ratelimit

import (
	"context"
	"errors"
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

func SetRateLimitConfigs(in map[string]RateLimitConfig) {
	if in != nil {
		modelRouteRateLimitConfigs = in
	}
	modelRouteRateLimitConfigs = map[string]RateLimitConfig{
		"gpt-3.5-turbo": {
			strategy: v1alpha1.RateLimitConfig_API_KEY,
			count:    defaultRateLimit,
			window:   defaultWindowInSeconds * time.Second,
		},
		"gpt-4": {
			strategy: v1alpha1.RateLimitConfig_API_KEY,
			count:    defaultRateLimit,
			window:   defaultWindowInSeconds * time.Second,
		},
	}
}

const (
	defaultRateLimit       = 100
	defaultWindowInSeconds = 60
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
		config:       config,
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

	config       RateLimitConfig
	apiKeyBucket map[string][]time.Time
	userBucket   map[string][]time.Time
	mu           sync.Mutex
}

func (rl *RateLimiter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	apiKey := rMeta.AuthInfo.GetApiKeyId()
	userName := rMeta.AuthInfo.GetUserId()

	modelName := request.GetModel()
	if cfg, ok := modelRouteRateLimitConfigs[modelName]; ok {
		rl.config = cfg
	}

	if !rl.AllowRequest(apiKey, userName) {
		return filters.NewFailed(errors.New("rate limit exceeded"))
	}

	return filters.NewOK()
}

func (rl *RateLimiter) checkBucket(bucket map[string][]time.Time, key string) bool {
	internal := rl.config.window
	if internal == 0 {
		internal = time.Second
	}
	now := time.Now()
	windowStart := now.Add(-internal)
	records := bucket[key]
	var validRecords []time.Time

	for _, record := range records {
		if record.After(windowStart) {
			validRecords = append(validRecords, record)
		}
	}
	bucket[key] = validRecords

	if len(validRecords) >= rl.config.count {
		return false
	}

	bucket[key] = append(bucket[key], now)

	return true
}

func (rl *RateLimiter) AllowRequest(apiKey, userName string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	switch rl.config.strategy {
	case v1alpha1.RateLimitConfig_API_KEY:
		return rl.checkBucket(rl.apiKeyBucket, apiKey)
	case v1alpha1.RateLimitConfig_USER:
		return rl.checkBucket(rl.userBucket, userName)
	case v1alpha1.RateLimitConfig_API_KEY_AND_USER:
		if !rl.checkBucket(rl.apiKeyBucket, apiKey) {
			return false
		}

		return rl.checkBucket(rl.userBucket, userName)
	default:
		return true
	}
}
