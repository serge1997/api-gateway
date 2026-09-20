package ratelimit

import "testing"

type RateLimiter interface {
	Allow() error
	ClearCacheForTest(t *testing.T)
}
