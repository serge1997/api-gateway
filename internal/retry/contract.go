package retry

import (
	"errors"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

var ErrBackoffAttemptsExceeded error = errors.New("retry backoff attempts exceeded")

type Op func() result.Result[httpresponse.HttpResponse]
type RetryBackoffStrategy interface {
	Execute(f Op, cb *circuitbreaker.CircuitBreaker) result.Result[httpresponse.HttpResponse]
	NextInterval() time.Duration
	Config() *RetryBackoffConfig
}
