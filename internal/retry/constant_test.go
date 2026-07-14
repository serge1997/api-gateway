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

var retryBackoff = retry.New(&retry.RetryBackoffConfig{
	Attempts: 3,
	Backoff:  "constant",
	Delay:    time.Millisecond * 100,
})
var cbConfig = &circuitbreaker.Config{
	FailureThreshold: 5,
	RetryTimeout:     "500ms",
}

func TestBackoffRetryIsConstant(t *testing.T) {
	if !retryBackoff.Config().Backoff.IsConstant() {
		t.Errorf("expect %v got %v", "constant", retryBackoff.Config().Backoff)
	}
}

func TestReachMaxAttempts(t *testing.T) {
	cb, _ := circuitbreaker.New(cbConfig)
	execResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("some data missing."), http.StatusServiceUnavailable))
	}, cb)
	if !strings.Contains(execResult.Value().Message, retry.ErrBackoffAttemptsExceeded.Error()) {
		t.Errorf("expect: %s %s got: %s", retry.ErrBackoffAttemptsExceeded.Error(), "some data missing", execResult.Value().Message)
	}
	if execResult.Value().Status != http.StatusServiceUnavailable {
		t.Errorf("expect status %d got %d", http.StatusServiceUnavailable, execResult.Value().Status)
	}
}

func TestMustReturnCbError(t *testing.T) {
	cb, _ := circuitbreaker.New(cbConfig)
	var execResult result.Result[httpresponse.HttpResponse]
	for range 10 {
		execResult = retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("internal server err"), http.StatusServiceUnavailable))
		}, cb)
		if cb.IsOpen() {
			break
		}
	}
	if !strings.Contains(execResult.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expect: %s %s got %s", circuitbreaker.ErrUnacessibleService.Error(), "internal server err", execResult.Value().Message)
	}
}

func TestExecuteMustSuccess(t *testing.T) {
	cb, _ := circuitbreaker.New(cbConfig)
	execResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Ok(httpresponse.SuccessResponse(nil, http.StatusOK, ""))
	}, cb)
	if !execResult.IsSuccess() {
		t.Errorf("expect Execute must success, got %v", execResult.IsSuccess())
	}
}
