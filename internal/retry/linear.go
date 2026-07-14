package retry

import (
	"fmt"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type linearBackoff struct {
	retryBackoff
}

func (l *linearBackoff) Execute(op Op, cb *circuitbreaker.CircuitBreaker) result.Result[httpresponse.HttpResponse] {
	var linearResult result.Result[httpresponse.HttpResponse]
	for at := 1; at <= int(l.config.Attempts); at++ {
		if cb != nil && cb.IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		linearResult = op()
		if linearResult.IsSuccess() {
			l.recordCbSuccess(cb)
			return linearResult
		}
		l.recordCbFailure(cb)
		if at == int(l.config.Attempts) {
			break
		}
		time.Sleep(l.NextInterval())
		continue
	}
	return result.Fail(httpresponse.FailResponse(
		fmt.Errorf("%s. reason: %s", ErrBackoffAttemptsExceeded, linearResult.Value().Message),
		http.StatusServiceUnavailable),
	)
}

func (l *linearBackoff) NextInterval() time.Duration {
	if l.interval < l.config.Delay {
		l.interval = l.config.Delay
	} else {
		l.interval += l.config.Delay
	}
	return l.interval
}
