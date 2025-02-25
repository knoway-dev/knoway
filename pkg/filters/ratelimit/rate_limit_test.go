package ratelimit

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	routev1alpha1 "knoway.dev/api/route/v1alpha1"
)

func TestCheckBucket(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		shards:    make([]*rateLimitShard, numShards),
		numShards: numShards,
		cancel:    cancel,
	}

	// Initialize shards
	for i := range numShards {
		rl.shards[i] = &rateLimitShard{
			buckets:        make(map[string]*tokenBucket),
			lastAccessTime: make(map[string]time.Time),
		}
	}

	type request struct {
		delay    time.Duration
		expected bool
	}

	tests := []struct {
		name     string
		key      string
		window   time.Duration
		limit    int
		requests []request
	}{
		{
			name:   "2 requests per second",
			key:    "test1",
			window: time.Second,
			limit:  2,
			requests: []request{
				{delay: 0, expected: true},                       // First request
				{delay: 0, expected: true},                       // Second request
				{delay: 0, expected: false},                      // Third request - should fail
				{delay: 500 * time.Millisecond, expected: false}, // Not enough tokens yet
				{delay: 500 * time.Millisecond, expected: true},  // Tokens replenished
				{delay: 0, expected: true},                       // Second request in new window
				{delay: 0, expected: false},                      // Third request - should fail
			},
		},
		{
			name:   "20 requests per second burst",
			key:    "test2",
			window: time.Second,
			limit:  20,
			requests: []request{
				// First burst of 20 requests
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				{delay: 0, expected: true},
				// Verify rate limiting
				{delay: 0, expected: false},                      // 21st request - should fail
				{delay: 500 * time.Millisecond, expected: false}, // Half window - still failing
				{delay: 500 * time.Millisecond, expected: true},  // Full window - should succeed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, req := range tt.requests {
				if req.delay > 0 {
					time.Sleep(req.delay)
				}

				got := rl.checkBucket(tt.key, tt.window, tt.limit)
				if got != req.expected {
					t.Errorf("Request #%d: got %v, want %v", i+1, got, req.expected)
				}
			}
		})
	}
}

func TestRateLimiter_CheckRateLimitPolicies(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		userName string
		route    string
		policy   []*routev1alpha1.RateLimitPolicy
		requests int
		expected []bool
	}{
		{
			name:     "basic api key limit - 2 requests per minute",
			apiKey:   "key1",
			userName: "user1",
			route:    "route1",
			policy: []*routev1alpha1.RateLimitPolicy{
				{
					BasedOn:  routev1alpha1.RateLimitBaseOn_API_KEY,
					Limit:    2,
					Duration: durationpb.New(60 * time.Second), // 1 minute
				},
			},
			requests: 3,
			expected: []bool{true, true, false}, // First 2 allowed, 3rd rejected
		},
		{
			name:     "basic user id limit - 2 requests per minute",
			apiKey:   "key1",
			userName: "user1",
			route:    "route1",
			policy: []*routev1alpha1.RateLimitPolicy{
				{
					BasedOn:  routev1alpha1.RateLimitBaseOn_USER_ID,
					Limit:    2,
					Duration: durationpb.New(60 * time.Second), // 1 minute
				},
			},
			requests: 3,
			expected: []bool{true, true, false}, // First 2 allowed, 3rd rejected
		},
		{
			name:     "api key limit with match exact",
			apiKey:   "special-key",
			userName: "user1",
			route:    "route1",
			policy: []*routev1alpha1.RateLimitPolicy{
				{
					Match: &routev1alpha1.StringMatch{
						Match: &routev1alpha1.StringMatch_Exact{
							Exact: "special-key",
						},
					},
					BasedOn:  routev1alpha1.RateLimitBaseOn_API_KEY,
					Limit:    5,
					Duration: durationpb.New(60 * time.Second),
				},
			},
			requests: 6,
			expected: []bool{true, true, true, true, true, false}, // First 5 allowed, 6th rejected
		},
		{
			name:     "api key limit with match prefix",
			apiKey:   "key-vip-123",
			userName: "user1",
			route:    "route1",
			policy: []*routev1alpha1.RateLimitPolicy{
				{
					Match: &routev1alpha1.StringMatch{
						Match: &routev1alpha1.StringMatch_Prefix{
							Prefix: "key-vip-",
						},
					},
					BasedOn:  routev1alpha1.RateLimitBaseOn_API_KEY,
					Limit:    3,
					Duration: durationpb.New(60 * time.Second),
				},
			},
			requests: 4,
			expected: []bool{true, true, true, false}, // First 3 allowed, 4th rejected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cancel := context.WithCancel(context.Background())

			rl := &RateLimiter{
				shards:    make([]*rateLimitShard, numShards),
				numShards: numShards,
				cancel:    cancel,
			}

			for i := range numShards {
				rl.shards[i] = &rateLimitShard{
					buckets:        make(map[string]*tokenBucket),
					lastAccessTime: make(map[string]time.Time),
				}
			}

			for i := range tt.requests {
				got := rl.CheckRateLimitPolicies(tt.apiKey, tt.userName, tt.route, tt.policy)
				if got != tt.expected[i] {
					t.Errorf("Request %d = %v, want %v", i+1, got, tt.expected[i])
				}
			}
		})
	}
}
