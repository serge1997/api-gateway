package retry

import (
	"fmt"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/contracts"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type exponentialBackoff struct {
	retryBackoff
}

func (e *exponentialBackoff) Execute(op Op, service contracts.ServiceRetry) result.Result[httpresponse.HttpResponse] {
	var expResult result.Result[httpresponse.HttpResponse]
	for at := 1; at <= int(e.config.Attempt()); at++ {
		if !service.CbIsNil() && service.Cb().IsOpen() {
			return result.Fail(httpresponse.FailResponse(
				circuitbreaker.ErrUnacessibleService,
				http.StatusServiceUnavailable),
			)
		}
		//the proxy check rate limit so here
		// do the ckeck after the first attempts
		if at > 1 {
			if err := service.Allow(); err != nil {
				return result.Fail(httpresponse.FailResponse(err, http.StatusTooManyRequests))
			}
		}
		expResult = op()
		if expResult.IsSuccess() {
			e.recordCbSuccess(service.Cb())
			return expResult
		}
		e.recordCbFailure(service.Cb())
		if at == int(e.config.Attempts) {
			break
		}
		time.Sleep(e.NextInterval())
	}
	return result.Fail(httpresponse.FailResponse(
		fmt.Errorf("%s. reason: %s", ErrBackoffAttemptsExceeded, expResult.Value().Message),
		http.StatusServiceUnavailable),
	)
}

func (e *exponentialBackoff) NextInterval() time.Duration {
	if e.interval < e.config.Delay {
		e.interval = e.config.Delay
	} else {
		e.interval *= 2
	}
	return e.interval
}
