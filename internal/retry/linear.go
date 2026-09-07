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

type linearBackoff struct {
	retryBackoff
}

func (l *linearBackoff) Execute(op Op, service contracts.ServiceRetry) result.Result[httpresponse.HttpResponse] {
	var linearResult result.Result[httpresponse.HttpResponse]
	for at := 1; at <= int(l.config.Attempt()); at++ {
		if !service.CbIsNil() && service.Cb().IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		if err := service.Allow(); err != nil {
			return result.Fail(httpresponse.FailResponse(err, http.StatusTooManyRequests))
		}
		linearResult = op()
		if linearResult.IsSuccess() {
			l.recordCbSuccess(service.Cb())
			return linearResult
		}
		l.recordCbFailure(service.Cb())
		if at == int(l.config.Attempts) {
			break
		}
		time.Sleep(l.NextInterval())
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

func (l *linearBackoff) Interval() time.Duration {
	return l.interval
}
