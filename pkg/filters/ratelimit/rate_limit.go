package ratelimit

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"

	"knoway.dev/api/filters/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/protoutils"

	"google.golang.org/protobuf/types/known/anypb"
)

const (
	cleanupInterval    = 30 * time.Minute
	cleanupThreshold   = 24 * time.Hour
	defaultDuration    = 1 * time.Minute
	numShards          = 64    // Number of shards, must be power of 2
	maxBucketsPerShard = 10000 // Maximum buckets per shard
	precision          = 1000  // Precision for fixed-point arithmetic
)

type tokenBucket struct {
	tokens     atomic.Int64 // Store tokens * precision
	capacity   atomic.Int64 // Store capacity * precision
	rate       atomic.Int64 // Store rate * precision
	lastUpdate atomic.Int64
	oldLimit   atomic.Int64
}

type rateLimitShard struct {
	buckets        map[string]*tokenBucket
	lastAccessTime map[string]time.Time
	mu             sync.Mutex
}

type RateLimiter struct {
	filters.IsRequestFilter

	shards    []*rateLimitShard
	numShards int
	cancel    context.CancelFunc

	pluginPolicies []*v1alpha1.RateLimitPolicy
}

var _ filters.RequestFilter = (*RateLimiter)(nil)
var _ filters.OnCompletionRequestFilter = (*RateLimiter)(nil)
var _ filters.OnImageGenerationsRequestFilter = (*RateLimiter)(nil)

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	rCfg, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		shards:    make([]*rateLimitShard, numShards),
		numShards: numShards,
		cancel:    cancel,

		pluginPolicies: rCfg.GetPolicies(),
	}

	// init shards
	for i := range numShards {
		rl.shards[i] = &rateLimitShard{
			buckets:        make(map[string]*tokenBucket),
			lastAccessTime: make(map[string]time.Time),
		}
	}

	// start cleanup
	go rl.cleanupLoop(ctx)

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStop: func(ctx context.Context) error {
			rl.cancel()
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

func (rl *RateLimiter) getShard(key string) *rateLimitShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIndex := h.Sum32() % uint32(rl.numShards)

	return rl.shards[shardIndex]
}

// Cleanup old keys that haven't been accessed for more than 24 hours
func (rl *RateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-ctx.Done():
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now()
	threshold := now.Add(-cleanupThreshold)

	// Clean each shard
	for _, shard := range rl.shards {
		shard.mu.Lock()
		for key, lastAccess := range shard.lastAccessTime {
			if lastAccess.Before(threshold) {
				delete(shard.buckets, key)
				delete(shard.lastAccessTime, key)
			}
		}
		shard.mu.Unlock()
	}
}

func (rl *RateLimiter) checkBucket(key string, window time.Duration, limit int) bool {
	if limit == 0 {
		return true
	}

	if window.Seconds() == 0 {
		window = defaultDuration
	}

	shard := rl.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	now := time.Now()
	shard.lastAccessTime[key] = now

	newCapacity := int64(limit * precision)
	rateInternal := window.Seconds()
	newRate := int64(float64(newCapacity) / rateInternal)

	bucket, exists := shard.buckets[key]
	if !exists {
		// check the maximum number of buckets for this shard
		if len(shard.buckets) >= maxBucketsPerShard {
			var oldestKey string
			oldestTime := now

			for k, t := range shard.lastAccessTime {
				if t.Before(oldestTime) {
					oldestTime = t
					oldestKey = k
				}
			}

			delete(shard.buckets, oldestKey)
			delete(shard.lastAccessTime, oldestKey)
		}

		// init new token bucket with fixed-point precision
		bucket = &tokenBucket{}
		bucket.capacity.Store(newCapacity)
		bucket.rate.Store(newRate)
		bucket.tokens.Store(newCapacity)
		bucket.lastUpdate.Store(now.UnixNano())
		bucket.oldLimit.Store(int64(limit))
		shard.buckets[key] = bucket
	} else if bucket.oldLimit.Load() != int64(limit) {
		bucket.oldLimit.Store(int64(limit))
		bucket.capacity.Store(newCapacity)
		bucket.rate.Store(newRate)
	}

	// calculate tokens to add based on time elapsed
	lastUpdateNano := bucket.lastUpdate.Load()
	lastUpdate := time.Unix(0, lastUpdateNano)
	elapsed := now.Sub(lastUpdate).Seconds()
	elapsedInt := int64(elapsed)

	tokensToAdd := elapsedInt * bucket.rate.Load()
	if tokensToAdd > 0 {
		newTokens := bucket.tokens.Load() + tokensToAdd
		if newTokens > bucket.capacity.Load() {
			newTokens = bucket.capacity.Load()
		}

		bucket.tokens.Store(newTokens)
		bucket.lastUpdate.Store(now.UnixNano())
	}

	if bucket.tokens.Load() >= precision {
		bucket.tokens.Add(-precision)
		return true
	}

	return false
}

func buildKey(baseOn v1alpha1.RateLimitBaseOn, value string, routeName string) string {
	return fmt.Sprintf("%s:%s:%s", baseOn, value, routeName)
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
		return filters.NewOK()
	}
	if rMeta.MatchRoute == nil || rMeta.MatchRoute.GetRouteConfig() == nil {
		return filters.NewOK()
	}

	route := rMeta.MatchRoute
	routeName := route.GetRouteConfig().GetName()

	var fPolicy *v1alpha1.RateLimitPolicy
	if routeName == "" {
		routeName = request.GetModel()
	}

	var rCfg *v1alpha1.RateLimitConfig

	for _, f := range route.GetRouteConfig().GetFilters() {
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

	if !rl.allowRequest(apiKey, userName, routeName, fPolicy) {
		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
}

func (rl *RateLimiter) allowRequest(apiKey, userName string, routeName string, policy *v1alpha1.RateLimitPolicy) bool {
	if policy == nil {
		return true
	}

	var value string

	switch policy.GetBasedOn() {
	case v1alpha1.RateLimitBaseOn_API_KEY:
		value = apiKey
	case v1alpha1.RateLimitBaseOn_USER_ID:
		value = userName
	case v1alpha1.RateLimitBaseOn_RATE_LIMIT_BASE_ON_UNSPECIFIED:
		return true
	default:
		return true
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
		return true
	}

	// disabled limit
	if policy.GetLimit() == 0 {
		return true
	}

	duration := policy.GetDuration().AsDuration()
	if duration == 0 {
		duration = defaultDuration
	}

	key := buildKey(policy.GetBasedOn(), value, routeName)

	return rl.checkBucket(key, duration, int(policy.GetLimit()))
}
