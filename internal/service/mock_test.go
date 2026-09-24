package service

import (
	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
)

var mockService = Service{
	Name:        "posts",
	Target:      "http://app.post",
	Middlewares: []string{"logger", "auth"},
}

func mockServiceWithConfigurableTyple() Service {
	mockService.CbConfig = &circuitbreaker.Config{
		FailureThreshold: 5,
		RetryTimeout:     "500ms",
		Before:           []string{"logger"},
	}
	mockService.RateLimits = []*ratelimit.Config{
		{
			Type:     "bucket",
			Interval: "10s",
		},
	}
	mockService.RetryBackoff = &retry.RetryBackoffConfig{
		Attempts: 5,
		Backoff:  "constant",
	}
	return mockService
}
