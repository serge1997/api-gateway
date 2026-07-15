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

var exponentialRetryBackoff = retry.New(&retry.RetryBackoffConfig{
	Attempts: 4,
	Backoff:  "exponential",
	Delay:    time.Millisecond * 100,
})
var exponentialCbConfig = &circuitbreaker.Config{
	FailureThreshold: 3,
	RetryTimeout:     "60s",
}

func TestRetryBackoffIsExponential(t *testing.T) {
	if !exponentialRetryBackoff.Config().Backoff.IsExponential() {
		t.Errorf("expect %v got %v", "exponential", exponentialRetryBackoff.Config().Backoff)
	}
}
func TestExponentialCumulateInterval(t *testing.T) {
	expected := exponentialRetryBackoff.Config().Delay * time.Duration(exponentialRetryBackoff.Config().Attempt())
	cb, _ := circuitbreaker.New(exponentialCbConfig)
	res := exponentialRetryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("bad gateway"), http.StatusBadGateway))
	}, cb)
	if expected != exponentialRetryBackoff.Interval() {
		t.Errorf("cumulate interval must be %v got %v. details: %v", expected, exponentialRetryBackoff.Interval(), res.Value().Message)
	}
}

func TestExponentialMustReturnCbError(t *testing.T) {
	cb, _ := circuitbreaker.New(exponentialCbConfig)
	var res result.Result[httpresponse.HttpResponse]
	for range 5 {
		res = exponentialRetryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("unknow error"), 401))
		}, cb)
		if cb.IsOpen() {
			break
		}
	}
	if !strings.Contains(res.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expected circuit breaker err. got: %s", res.Value().Message)
	}
}
