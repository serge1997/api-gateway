package retry

import (
	"fmt"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
)

type retryBackoff struct {
	config   *RetryBackoffConfig
	interval time.Duration
}

func New(conf *RetryBackoffConfig) RetryBackoffStrategy {
	fmt.Println(conf.Backoff)
	backoff := retryBackoff{config: conf}
	if conf.Backoff.IsConstant() {
		return &constantBackoff{backoff}
	}
	if conf.Backoff.IsExponential() {
		return &exponentialBackoff{backoff}
	}
	if conf.Backoff.IsLinear() {
		return &linearBackoff{backoff}
	}

	if conf.Backoff.IsJitter() {
		return &jitterBackoff{backoff}
	}

	return nil
}

func (r *retryBackoff) recordCbFailure(cb *circuitbreaker.CircuitBreaker) {
	if cb != nil {
		cb.RecordFailure()
	}
}
func (r *retryBackoff) recordCbSuccess(cb *circuitbreaker.CircuitBreaker) {
	if cb != nil {
		cb.RecordSuccess()
	}
}

func (r *retryBackoff) Config() *RetryBackoffConfig {
	return r.config
}
