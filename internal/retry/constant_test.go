package retry_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/retry"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

func TestBackoffRetryIsConstant(t *testing.T) {
	if !retryBackoff.Config().Backoff.IsConstant() {
		t.Errorf("expect backof type %v got %v", "constant", retryBackoff.Config().Backoff)
	}
}

func TestReachMaxAttempts(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	execResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("some data missing."), http.StatusServiceUnavailable))
	}, service)
	if !strings.Contains(execResult.Value().Message, retry.ErrBackoffAttemptsExceeded.Error()) {
		t.Errorf("expect: %s %s got: %s", retry.ErrBackoffAttemptsExceeded.Error(), "some data missing", execResult.Value().Message)
	}
	if execResult.Value().Status != http.StatusServiceUnavailable {
		t.Errorf("expect status %d got %d", http.StatusServiceUnavailable, execResult.Value().Status)
	}
}

func TestMustReturnCbError(t *testing.T) {
	retryBackoff.Config().Backoff = "constant"
	retryBackoff = retry.New(retryBackoff.Config())
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	var execResult result.Result[httpresponse.HttpResponse]
	for range 10 {
		execResult = retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("internal server err"), http.StatusServiceUnavailable))
		}, service)
	}
	if !strings.Contains(execResult.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expect: %s %s got %s", circuitbreaker.ErrUnacessibleService.Error(), "internal server err", execResult.Value().Message)
	}
}

func TestExecuteMustSuccess(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	execResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Ok(httpresponse.SuccessResponse(nil, http.StatusOK, ""))
	}, service)
	if !execResult.IsSuccess() {
		t.Errorf("expect Execute must success, got %v", execResult.IsSuccess())
	}
}

func TestAttempsIsDefaultValue(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	retryBackoff.Config().Attempts = 0
	service := mockService{
		cb: cb,
	}
	execResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Ok(httpresponse.SuccessResponse(nil, http.StatusOK, ""))
	}, service)
	if !execResult.IsSuccess() {
		t.Errorf("expect Execute must success, got %v", execResult.Value().Message)
	}
	if retryBackoff.Config().Attempt() != retry.DefaultAttempts() {
		t.Errorf("expect attempt to be %d, got %v", retry.DefaultAttempts(), retryBackoff.Config().Attempt())
	}
	retryBackoff.Config().Attempts = 3
}
