package retry_test

import (
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/retry"
)

var cbConfig = &circuitbreaker.Config{
	FailureThreshold: 5,
	RetryTimeout:     "500ms",
}

type mockService struct {
	cb *circuitbreaker.CircuitBreaker
}

func (s mockService) Cb() *circuitbreaker.CircuitBreaker {
	return s.cb
}
func (s mockService) CbIsNil() bool {
	return s.cb == nil
}
func (s mockService) Allow() error {
	return nil
}

var retryBackoff = retry.New(&retry.RetryBackoffConfig{
	Attempts: 5,
	Backoff:  "constant",
	Delay:    time.Millisecond * 100,
})
