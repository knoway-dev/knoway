package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	_, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	return &RateLimiter{
		apiKeyBucket: make(map[string][]time.Time),
		userBucket:   make(map[string][]time.Time),
	}, nil
}

var _ filters.RequestFilter = (*RateLimiter)(nil)
var _ filters.OnCompletionRequestFilter = (*RateLimiter)(nil)

type RateLimiter struct {
	filters.IsRequestFilter

	apiKeyBucket map[string][]time.Time
	userBucket   map[string][]time.Time
	mu           sync.Mutex
}

func (rl *RateLimiter) OnCompletionRequest(ctx context.Context, request object.LLMRequest, sourceHTTPRequest *http.Request) filters.RequestFilterResult {
	rMeta := metadata.RequestMetadataFromCtx(ctx)
	apiKey := rMeta.AuthInfo.GetApiKeyId()
	userName := rMeta.AuthInfo.GetUserId()

	if apiKey == "" && userName == "" {
		return filters.NewOK()
	}

	rCfg := rMeta.MatchRoute
	if rCfg == nil || rCfg.GetRateLimitPolicy() == nil {
		return filters.NewOK()
	}

	// Check whitelists before rate limiting
	for _, whitelistedAPIKey := range rCfg.GetRateLimitPolicy().GetApiKeyWhitelist() {
		if apiKey == whitelistedAPIKey {
			return filters.NewOK()
		}
	}

	for _, whitelistedUser := range rCfg.GetRateLimitPolicy().GetUserWhitelist() {
		if userName == whitelistedUser {
			return filters.NewOK()
		}
	}

	if !rl.AllowRequestWithConfig(apiKey, userName, rCfg.GetRateLimitPolicy()) {
		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
}

func (rl *RateLimiter) checkBucket(bucket map[string][]time.Time, key string, window time.Duration, limit int32) bool {
	if window.Seconds() == 0 {
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

	if len(validRecords) >= int(limit) {
		return false
	}

	bucket[key] = append(bucket[key], now)

	return true
}

func (rl *RateLimiter) AllowRequestWithConfig(apiKey, userName string, config *routev1alpha1.RateLimitPolicy) bool {
	if config == nil {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if config.GetLimit() == 0 {
		return true
	}

	dur := config.GetDuration().AsDuration()
	// setting default
	if dur == 0 {
		const defaultRateLimitDuration = 300 * time.Second // 5 minutes
		dur = durationpb.New(defaultRateLimitDuration).AsDuration()
	}

	switch config.GetBaseOn() {
	case routev1alpha1.RateLimitBaseOn_API_KEY:
		return rl.checkBucket(rl.apiKeyBucket, apiKey, dur, config.GetLimit())
	case routev1alpha1.RateLimitBaseOn_USER:
		return rl.checkBucket(rl.userBucket, userName, dur, config.GetLimit())
	default:
		return true
	}
}
