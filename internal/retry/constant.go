package retry

import (
	"fmt"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type constantBackoff struct {
	retryBackoff
}

func (c *constantBackoff) Execute(op Op, cb *circuitbreaker.CircuitBreaker) result.Result[httpresponse.HttpResponse] {
	var consResult result.Result[httpresponse.HttpResponse]
	for at := 1; at < int(c.config.Attempt()); at++ {
		if cb != nil && cb.IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		consResult = op()
		if consResult.IsSuccess() {
			c.recordCbSuccess(cb)
			return consResult
		}
		c.recordCbFailure(cb)
		if at == int(c.config.Attempts) {
			break
		}
		time.Sleep(c.config.Delay)
	}
	return result.Fail(httpresponse.FailResponse(
		fmt.Errorf("%s. reason: %s", ErrBackoffAttemptsExceeded, consResult.Value().Message),
		http.StatusServiceUnavailable),
	)
}

func (c *constantBackoff) NextInterval() time.Duration {
	return time.Second
}
