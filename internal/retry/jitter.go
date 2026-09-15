package retry

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/contracts"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

var JitterRandMaxMs = 9

type jitterBackoff struct {
	retryBackoff
}

func (j *jitterBackoff) Execute(op Op, service contracts.ServiceRetry) result.Result[httpresponse.HttpResponse] {
	var jitResult result.Result[httpresponse.HttpResponse]
	for at := 1; at <= int(j.config.Attempt()); at++ {
		if !service.CbIsNil() && service.Cb().IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		if at > 1 {
			if err := service.Allow(); err != nil {
				return result.Fail(httpresponse.FailResponse(err, http.StatusTooManyRequests))
			}
		}
		jitResult = op()
		if jitResult.IsSuccess() {
			j.recordCbSuccess(service.Cb())
			return jitResult
		}
		j.recordCbFailure(service.Cb())
		if at == int(j.config.Attempts) {
			break
		}
		time.Sleep(j.NextInterval())
	}
	return result.Fail(httpresponse.FailResponse(
		fmt.Errorf("%s. reason: %s", ErrBackoffAttemptsExceeded, jitResult.Value().Message),
		http.StatusServiceUnavailable),
	)
}

func (j *jitterBackoff) NextInterval() time.Duration {
	randMs := rand.Intn(JitterRandMaxMs) + 1
	delay := time.Millisecond * time.Duration(randMs*100)
	return delay
}
