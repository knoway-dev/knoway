package ratelimit

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"knoway.dev/api/filters/v1alpha1"
	routev1alpha1 "knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/bootkit"
	"knoway.dev/pkg/filters"
	"knoway.dev/pkg/metadata"
	"knoway.dev/pkg/object"
	"knoway.dev/pkg/protoutils"

	"google.golang.org/protobuf/types/known/anypb"
)

const (
	cleanupInterval    = 30 * time.Minute
	cleanupThreshold   = 24 * time.Hour
	defaultDuration    = 5 * time.Minute
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
	stopCh    chan struct{}
}

func NewWithConfig(cfg *anypb.Any, lifecycle bootkit.LifeCycle) (filters.RequestFilter, error) {
	_, err := protoutils.FromAny(cfg, &v1alpha1.RateLimitConfig{})
	if err != nil {
		return nil, fmt.Errorf("invalid config type %T", cfg)
	}

	rl := &RateLimiter{
		shards:    make([]*rateLimitShard, numShards),
		numShards: numShards,
		stopCh:    make(chan struct{}),
	}

	// init shards
	//nolint:intrange
	for i := 0; i < numShards; i++ {
		rl.shards[i] = &rateLimitShard{
			buckets:        make(map[string]*tokenBucket),
			lastAccessTime: make(map[string]time.Time),
		}
	}

	// start cleanup
	go rl.cleanupLoop()

	lifecycle.Append(bootkit.LifeCycleHook{
		OnStop: func(ctx context.Context) error {
			close(rl.stopCh)
			return nil
		},
	})

	return rl, nil
}

var _ filters.RequestFilter = (*RateLimiter)(nil)
var _ filters.OnCompletionRequestFilter = (*RateLimiter)(nil)

func (rl *RateLimiter) getShard(key string) *rateLimitShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIndex := h.Sum32() % uint32(rl.numShards)

	return rl.shards[shardIndex]
}

// Cleanup old keys that haven't been accessed for more than 24 hours
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
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

	routeName := rCfg.GetName()

	if routeName == "" {
		return filters.NewOK()
	}

	if !rl.AllowRequestWithConfig(apiKey, userName, routeName, rCfg.GetRateLimitPolicy()) {
		return filters.NewFailed(object.NewErrorRateLimitExceeded())
	}

	return filters.NewOK()
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

func buildKey(baseOn routev1alpha1.RateLimitBaseOn, value string, routeName string) string {
	return fmt.Sprintf("%s:%s:%s", baseOn, value, routeName)
}

func (rl *RateLimiter) AllowRequestWithConfig(apiKey, userName string, routeName string, config *routev1alpha1.RateLimitPolicy) bool {
	if config == nil {
		return true
	}

	// check advanced limits first
	for _, advanceLimit := range config.GetAdvanceLimits() {
		on := routev1alpha1.RateLimitBaseOn_APIKey
		value := ""

		for _, obj := range advanceLimit.GetObjects() {
			switch obj.GetBaseOn() {
			case routev1alpha1.RateLimitBaseOn_APIKey:
				if apiKey == obj.GetValue() {
					value = apiKey
					break
				}
			case routev1alpha1.RateLimitBaseOn_User:
				if userName == obj.GetValue() {
					on = routev1alpha1.RateLimitBaseOn_User
					value = userName

					break
				}
			}
		}

		if value == "" {
			continue
		}

		// disabled limit
		if advanceLimit.GetLimit() == 0 {
			return true
		}

		key := buildKey(on, value, routeName)

		return rl.checkBucket(key, advanceLimit.GetDuration().AsDuration(), int(advanceLimit.GetLimit()))
	}

	// disabled limit
	if config.GetLimit() == 0 {
		return true
	}

	duration := config.GetDuration().AsDuration()
	if duration == 0 {
		duration = defaultDuration
	}

	var value string

	switch config.GetBaseOn() {
	case routev1alpha1.RateLimitBaseOn_APIKey:
		value = apiKey
	case routev1alpha1.RateLimitBaseOn_User:
		value = userName
	}

	key := buildKey(config.GetBaseOn(), value, routeName)
	allowed := rl.checkBucket(key, duration, int(config.GetLimit()))

	return allowed
}
