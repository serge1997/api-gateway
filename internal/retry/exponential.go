package retry

import (
	"fmt"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type exponentialBackoff struct {
	retryBackoff
}

func (e *exponentialBackoff) Execute(op Op, cb *circuitbreaker.CircuitBreaker) result.Result[httpresponse.HttpResponse] {
	var expResult result.Result[httpresponse.HttpResponse]
	for at := 1; at <= int(e.config.Attempts); at++ {
		if cb != nil && cb.IsOpen() {
			return result.Fail(httpresponse.FailResponse(
				circuitbreaker.ErrUnacessibleService,
				http.StatusServiceUnavailable),
			)
		}
		expResult = op()
		if expResult.IsSuccess() {
			e.recordCbSuccess(cb)
			return expResult
		}
		e.recordCbFailure(cb)
		if at == int(e.config.Attempts) {
			break
		}
		time.Sleep(e.NextInterval())
	}
	return result.Fail(httpresponse.FailResponse(
		fmt.Errorf("%s. reason: %s", ErrBackoffAttemptsExceded, expResult.Value().Message),
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
