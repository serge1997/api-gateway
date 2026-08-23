package contracts

import (
	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
)

type Service interface {
	HasRetryBackoffConfigured() bool
	CbIsNil() bool
	RateLimitsIsNil() bool
	RetryIsNil() bool
	Cb() *circuitbreaker.CircuitBreaker
	RtLimiters() []ratelimit.RateLimiter
	Allow() error
}
