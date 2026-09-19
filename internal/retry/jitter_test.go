package retry_test

import (
	"fmt"
	"testing"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/retry"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

func TestRetryBackoffIsJitter(t *testing.T) {
	retryBackoff.Config().Backoff = "jitter"
	if !retryBackoff.Config().Backoff.IsJitter() {
		t.Errorf("expect %v got %v", "jitter", retryBackoff.Config().Backoff)
	}
}

func TestJitterBackoffRetry(t *testing.T) {
	retryBackoff = retry.New(retryBackoff.Config())
	var cb, _ = circuitbreaker.New(cbConfig)
	max := time.Millisecond * time.Duration(retry.JitterRandMaxMs*100)
	service := mockService{
		cb: cb,
	}
	retryBackoff.Execute(func() result.Result[httpresponse.HttpResponse] {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("bad gateway"), 502))
	}, service)
	if retryBackoff.Interval() > max {
		t.Errorf("cumulate interval must be less than %v got %v", max, retryBackoff.Interval())
	}
}
