package retry

import (
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
	for at := 1; at < int(c.config.Attempts); at++ {
		opResult := op()
		if opResult.IsSuccess() {
			c.recordCbSuccess(cb)
			return opResult
		}
		c.recordCbFailure(cb)
		if cb != nil && cb.IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		if at == int(c.config.Attempts) {
			break
		}
		time.Sleep(c.config.Delay)
		continue
	}
	return result.Fail(httpresponse.FailResponse(ErrBackoffAttemptsExceded, http.StatusServiceUnavailable))
}

func (c *constantBackoff) NextInterval() time.Duration {
	return time.Second
}
