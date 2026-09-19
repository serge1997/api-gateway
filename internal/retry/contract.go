package retry

import (
	"errors"
	"time"

	"github.com/serge1997/apigateway/internal/contracts"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

var ErrBackoffAttemptsExceeded error = errors.New("retry backoff attempts exceeded")

type Op func() result.Result[httpresponse.HttpResponse]
type RetryBackoffStrategy interface {
	Execute(f Op, service contracts.ServiceRetry) result.Result[httpresponse.HttpResponse]
	NextInterval() time.Duration
	Interval() time.Duration
	Config() *RetryBackoffConfig
}
