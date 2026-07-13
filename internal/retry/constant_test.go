package retry_test

import (
	"net/http"
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
		return result.Fail(httpresponse.FailResponse(retry.ErrBackoffAttemptsExceded, http.StatusServiceUnavailable))
	}, cb)
	if execResult.Value().Message != retry.ErrBackoffAttemptsExceded.Error() {
		t.Errorf("expect %s got %s", retry.ErrBackoffAttemptsExceded.Error(), execResult.Value().Message)
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
			return result.Fail(httpresponse.FailResponse(retry.ErrBackoffAttemptsExceded, http.StatusServiceUnavailable))
		}, cb)
		if cb.IsOpen() {
			break
		}
	}
	if execResult.Value().Message != circuitbreaker.ErrUnacessibleService.Error() {
		t.Errorf("expect %s got %s", circuitbreaker.ErrUnacessibleService.Error(), execResult.Value().Message)
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
