package retry_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/retry"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

var linearRetryBackoff = retry.New(&retry.RetryBackoffConfig{
	Attempts: 3,
	Backoff:  "linear",
	Delay:    time.Millisecond * 100,
})
var linearCbConfig = &circuitbreaker.Config{
	FailureThreshold: 5,
	RetryTimeout:     "500ms",
}

func TestRetryBackOffIsLinear(t *testing.T) {
	if !linearRetryBackoff.Config().Backoff.IsLinear() {
		t.Errorf("expect %v got %v", "linear", linearRetryBackoff.Config().Backoff)
	}
}
func TestLinearCumulateInterval(t *testing.T) {
	expectCumulate := linearRetryBackoff.Config().Delay * time.Duration(linearRetryBackoff.Config().Attempt()-1)
	cb, _ := circuitbreaker.New(linearCbConfig)
	_ = linearRetryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("internal server erro"), http.StatusInternalServerError))
	}, cb)
	if expectCumulate != linearRetryBackoff.Interval() {
		t.Errorf("cumulate interval must be %v got %v", expectCumulate, linearRetryBackoff.Interval())
	}
}

func TestLinarMustReturnCbError(t *testing.T) {
	cb, _ := circuitbreaker.New(linearCbConfig)
	var res result.Result[httpresponse.HttpResponse]
	for range 6 {
		res = linearRetryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("syntax error"), 501))
		}, cb)
		if cb.IsOpen() {
			break
		}
	}
	if !strings.Contains(res.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expected circuit breaker err. got: %s", res.Value().Message)
	}
}

func TestLinearMaxAttemdReached(t *testing.T) {
	cb, _ := circuitbreaker.New(linearCbConfig)
	var linear = retry.New(&retry.RetryBackoffConfig{
		Attempts: 3,
		Backoff:  "linear",
		Delay:    time.Millisecond * 100,
	})
	linearRetryResult := linear.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(
			fmt.Errorf("%s. reason: %s", retry.ErrBackoffAttemptsExceeded, "permission denied"),
			http.StatusInternalServerError),
		)
	}, cb)
	if !strings.Contains(linearRetryResult.Value().Message, retry.ErrBackoffAttemptsExceeded.Error()) {
		t.Errorf("expect: %s got: %s", retry.ErrBackoffAttemptsExceeded.Error(), linearRetryResult.Value().Message)
	}
}
