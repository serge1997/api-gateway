package ratelimit

type RateLimiter interface {
	Allow() error
}
