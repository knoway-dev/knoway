package ratelimit

import (
	"context"
	"hash/fnv"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type tokenBucket struct {
	tokens     atomic.Int64 // Store tokens * precision
	capacity   atomic.Int64 // Store capacity * precision
	rate       atomic.Int64 // Store rate * precision
	lastUpdate atomic.Int64
	oldLimit   atomic.Int64
	expireAt   atomic.Int64 // TTL expiration timestamp
}

type rateLimitShard struct {
	buckets        map[string]*tokenBucket
	lastAccessTime map[string]time.Time
	mu             sync.RWMutex // 使用读写锁提高并发性能
}

func (rl *RateLimiter) getShard(key string) *rateLimitShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIndex := h.Sum32() % uint32(rl.numShards)

	return rl.shards[shardIndex]
}

// checkBucketLocal Local rate limiting
func (rl *RateLimiter) checkBucketLocal(key string, window time.Duration, limit int) (bool, error) {
	shard := rl.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	now := time.Now()
	shard.lastAccessTime[key] = now

	// Calculate TTL - max(5min, 2x window)
	ttl := max(maxTTL, window*ttlRate)
	expireAt := now.Add(ttl).UnixNano()

	newCapacity := int64(limit * precision)
	newRate := int64(float64(newCapacity) / window.Seconds())

	bucket := shard.buckets[key]
	if bucket == nil {
		bucket = rl.initBucket(shard, key, limit, newCapacity, newRate, now, expireAt)
	} else {
		bucket = rl.updateBucket(shard, bucket, key, limit, newCapacity, newRate, now, expireAt)
	}

	return rl.tryConsume(bucket, now, key), nil
}

func (rl *RateLimiter) initBucket(shard *rateLimitShard, key string, limit int, newCapacity, newRate int64, now time.Time, expireAt int64) *tokenBucket {
	// check the maximum number of buckets for this shard
	if len(shard.buckets) >= maxBucketsPerShard {
		rl.evictOldestBucket(shard, now)
	}

	// init new token bucket with fixed-point precision
	bucket := &tokenBucket{}
	bucket.capacity.Store(newCapacity)
	bucket.rate.Store(newRate)
	bucket.tokens.Store(newCapacity)
	bucket.lastUpdate.Store(now.UnixNano())
	bucket.oldLimit.Store(int64(limit))
	bucket.expireAt.Store(expireAt)
	shard.buckets[key] = bucket

	slog.DebugContext(context.Background(), "created new token bucket", append(rl.logCommonAttrs(), slog.String("key", key), slog.Int("limit", limit))...)

	return bucket
}

func (rl *RateLimiter) updateBucket(shard *rateLimitShard, bucket *tokenBucket, key string, limit int, newCapacity, newRate int64, now time.Time, expireAt int64) *tokenBucket {
	// Check if bucket has expired
	if now.UnixNano() > bucket.expireAt.Load() {
		delete(shard.buckets, key)
		delete(shard.lastAccessTime, key)
		slog.DebugContext(context.Background(), "bucket expired, creating new one", append(rl.logCommonAttrs(), slog.String("key", key))...)

		return rl.initBucket(shard, key, limit, newCapacity, newRate, now, expireAt)
	}

	if bucket.oldLimit.Load() != int64(limit) {
		slog.DebugContext(context.Background(), "updating bucket limit", append(rl.logCommonAttrs(), slog.String("key", key), slog.Int64("oldLimit", bucket.oldLimit.Load()), slog.Int("newLimit", limit))...)
		bucket.oldLimit.Store(int64(limit))
		bucket.capacity.Store(newCapacity)
		bucket.rate.Store(newRate)
	}

	// Update TTL on access
	bucket.expireAt.Store(expireAt)

	return bucket
}

func (rl *RateLimiter) evictOldestBucket(shard *rateLimitShard, now time.Time) {
	var oldestKey string
	oldestTime := now

	for k, t := range shard.lastAccessTime {
		if t.Before(oldestTime) {
			oldestTime = t
			oldestKey = k
		}
	}

	slog.DebugContext(context.Background(), "removing oldest bucket to make space", append(rl.logCommonAttrs(), slog.String("key", oldestKey))...)
	delete(shard.buckets, oldestKey)
	delete(shard.lastAccessTime, oldestKey)
}

func (rl *RateLimiter) tryConsume(bucket *tokenBucket, now time.Time, key string) bool {
	// calculate tokens to add based on time elapsed
	lastUpdateNano := bucket.lastUpdate.Load()
	elapsed := now.Sub(time.Unix(0, lastUpdateNano)).Seconds()
	tokensToAdd := int64(elapsed) * bucket.rate.Load()

	if tokensToAdd > 0 {
		newTokens := min(bucket.tokens.Load()+tokensToAdd, bucket.capacity.Load())
		bucket.tokens.Store(newTokens)
		bucket.lastUpdate.Store(now.UnixNano())
	}

	if bucket.tokens.Load() >= precision {
		bucket.tokens.Add(-precision)
		return true
	}

	slog.DebugContext(context.Background(), "rate limit exceeded", append(rl.logCommonAttrs(), slog.String("key", key), slog.Int64("tokens", bucket.tokens.Load()), slog.Int("precision", precision))...)

	return false
}

// Cleanup old keys that haven't been accessed for more than 24 hours
func (rl *RateLimiter) cleanupLoop(ctx context.Context) {
	slog.DebugContext(context.Background(), "starting cleanup loop", append(rl.logCommonAttrs(), slog.Duration("interval", cleanupInterval))...)
	ticker := time.NewTicker(cleanupInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-ctx.Done():
			slog.InfoContext(context.Background(), "stopping cleanup loop", rl.logCommonAttrs()...)
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	slog.DebugContext(context.Background(), "cleaning up expired rate limit buckets", rl.logCommonAttrs()...)
	now := time.Now()

	// Clean each shard
	for i, shard := range rl.shards {
		shard.mu.Lock()
		beforeCount := len(shard.buckets)

		for key, bucket := range shard.buckets {
			if now.UnixNano() > bucket.expireAt.Load() {
				delete(shard.buckets, key)
				delete(shard.lastAccessTime, key)
			}
		}
		afterCount := len(shard.buckets)
		slog.DebugContext(context.Background(), "cleaned shard", append(rl.logCommonAttrs(), slog.Int("shard", i), slog.Int("removed", beforeCount-afterCount))...)
		shard.mu.Unlock()
	}
}
