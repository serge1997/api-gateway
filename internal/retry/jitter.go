package retry

import (
	"math/rand"
	"net/http"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type jitterBackoff struct {
	retryBackoff
}

func (j *jitterBackoff) Execute(op Op, cb *circuitbreaker.CircuitBreaker) result.Result[httpresponse.HttpResponse] {
	for at := 1; at <= int(j.config.Attempts); at++ {
		jitResult := op()
		if jitResult.IsSuccess() {
			j.recordCbSuccess(cb)
			return jitResult
		}
		j.recordCbFailure(cb)
		if cb != nil && cb.IsOpen() {
			return result.Fail(httpresponse.FailResponse(circuitbreaker.ErrUnacessibleService, http.StatusServiceUnavailable))
		}
		if at == int(j.config.Attempts) {
			break
		}
		time.Sleep(j.NextInterval())
		continue
	}
	return result.Fail(httpresponse.FailResponse(ErrBackoffAttemptsExceded, http.StatusServiceUnavailable))
}

func (j *jitterBackoff) NextInterval() time.Duration {
	randMs := rand.Intn(9) + 1
	delay := time.Millisecond * time.Duration(randMs*100)
	return delay
}
