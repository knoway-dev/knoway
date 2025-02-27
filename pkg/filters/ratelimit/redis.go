package ratelimit

import (
	"context"
	"log/slog"
	"strconv"
	"time"
)

//nolint:dupword
var redisRateLimitScript = `
-- KEYS[1]: rate limit key
-- ARGV[1]: limit (max tokens)
-- ARGV[2]: window in milliseconds
-- ARGV[3]: current timestamp in milliseconds
-- ARGV[4]: precision multiplier

local function init_bucket(limit, now)
    return {
        tokens = limit,
        last_update = now,
        limit = limit
    }
end

local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local precision = tonumber(ARGV[4])

-- Calculate bucket parameters
local capacity = limit * precision
local fill_rate = capacity / window_ms -- tokens per millisecond

-- Get or create bucket atomically
local bucket = redis.call('HGETALL', key)
local state = {}
if #bucket == 0 then
    state = init_bucket(capacity, now)
else
    -- Convert array to hash
    for i = 1, #bucket, 2 do
        state[bucket[i]] = tonumber(bucket[i + 1])
    end
    
    -- Handle rate limit changes
    if state.limit ~= capacity then
        state = init_bucket(capacity, now)
    end
end

-- Calculate available tokens
local elapsed_ms = now - state.last_update
local new_tokens = math.min(capacity, state.tokens + (elapsed_ms * fill_rate))

-- Attempt to consume token
local allowed = 0
if new_tokens >= precision then
    new_tokens = new_tokens - precision
    allowed = 1
end

-- Update bucket state
local ttl = math.max(300000, math.ceil(window_ms * 2)) -- Set TTL to max(5min, 2x window) for safety
redis.call('HMSET', key,
    'tokens', new_tokens,
    'last_update', now,
    'limit', capacity
)
redis.call('PEXPIRE', key, ttl)

return allowed
`

func (rl *RateLimiter) checkBucketRedis(key string, window time.Duration, limit int) (bool, error) {
	now := time.Now().UnixMilli() // 使用毫秒精度
	windowMs := window.Milliseconds()

	cmd := rl.redisClient.B().Eval().Script(redisRateLimitScript).
		Numkeys(1).
		Key(key).
		Arg(
			strconv.Itoa(limit),
			strconv.FormatInt(windowMs, 10),
			strconv.FormatInt(now, 10),
			strconv.Itoa(precision),
		).
		Build()

	result := rl.redisClient.Do(context.Background(), cmd)
	if err := result.NonRedisError(); err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "redis error", append(rl.logCommonAttrs(), slog.Any("error", err))...)
		return false, err
	}

	allowed, err := result.AsInt64()
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "failed to parse redis result", append(rl.logCommonAttrs(), slog.Any("error", err))...)
		return false, err
	}

	return allowed != 0, nil
}
