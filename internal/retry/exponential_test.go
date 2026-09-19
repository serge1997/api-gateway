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

func TestRetryBackoffIsExponential(t *testing.T) {
	retryBackoff.Config().Backoff = "exponential"
	if !retryBackoff.Config().Backoff.IsExponential() {
		t.Errorf("expect %v got %v", "exponential", retryBackoff.Config().Backoff)
	}
}
func TestExponentialBackoffRetryCumulateInterval(t *testing.T) {
	retryBackoff.Config().Backoff = "exponential"
	retryBackoff = retry.New(retryBackoff.Config())
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	expected := retryBackoff.Config().Delay
	attempts := int(retryBackoff.Config().Attempts - 1)
	for i := 1; i < attempts; i++ {
		expected *= 2
	}
	res := retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("bad gateway"), http.StatusBadGateway))
	}, service)
	if expected != retryBackoff.Interval() {
		t.Errorf("cumulate interval must be %v got %v. details: %v", expected, retryBackoff.Interval(), res.Value().Message)
	}
}

func TestExponentialBackoffRetryMustReturnCbError(t *testing.T) {
	var cb, _ = circuitbreaker.New(cbConfig)
	service := mockService{
		cb: cb,
	}
	var res result.Result[httpresponse.HttpResponse]
	for range 5 {
		res = retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
			return result.Fail(httpresponse.FailResponse(fmt.Errorf("unknow error"), 401))
		}, service)
		if service.cb.IsOpen() {
			break
		}
	}
	if !strings.Contains(res.Value().Message, circuitbreaker.ErrUnacessibleService.Error()) {
		t.Errorf("expected circuit breaker err. got: %s", res.Value().Message)
	}
}
