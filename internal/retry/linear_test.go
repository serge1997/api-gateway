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

func TestRetryBackOffIsLinear(t *testing.T) {
	retryBackoff.Config().Backoff = "linear"
	if !retryBackoff.Config().Backoff.IsLinear() {
		t.Errorf("expect %v got %v", "linear", retryBackoff.Config().Backoff)
	}
}
func TestLinearBackoffRetryCumulateInterval(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	retryBackoff.Config().Backoff = "linear"
	retryBackoff = retry.New(retryBackoff.Config())
	expectCumulate := retryBackoff.Config().Delay * time.Duration(retryBackoff.Config().Attempt()-1)
	_ = retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("internal server erro"), http.StatusInternalServerError))
	}, service)
	if expectCumulate != retryBackoff.Interval() {
		t.Errorf("cumulate interval must be %v got %v", expectCumulate, retryBackoff.Interval())
	}
}

func TestLinarBackoffRetryMustReturnCbError(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	retryBackoff.Config().Backoff = "linear"
	retryBackoff = retry.New(retryBackoff.Config())
	var res result.Result[httpresponse.HttpResponse]
	for range 6 {
		res = retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("syntax error"), 501))
		}, service)
		if service.cb.IsOpen() {
			break
		}
	}
	if !strings.Contains(res.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expected circuit breaker err. got: %s", res.Value().Message)
	}
}

func TestLinearBackoffRetryMaxAttemdReached(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	linearRetryResult := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(
			fmt.Errorf("%s. reason: %s", retry.ErrBackoffAttemptsExceeded, "permission denied"),
			http.StatusInternalServerError),
		)
	}, service)
	if !strings.Contains(linearRetryResult.Value().Message, retry.ErrBackoffAttemptsExceeded.Error()) {
		t.Errorf("expect: %s got: %s", retry.ErrBackoffAttemptsExceeded.Error(), linearRetryResult.Value().Message)
	}
}
