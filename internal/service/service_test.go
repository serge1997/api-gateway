package service

import (
	"testing"
	"time"
)

func TestServiceTimeout(t *testing.T) {
	t.Run("set default timeout value when config is not provided", func(t *testing.T) {
		expected := mockService.GetTimeout()
		if expected != defaultTimeout {
			t.Errorf("timeout must be %s got %s", defaultTimeout, mockService.Timeout)
		}
	})

	t.Run("service timeout must use provided value and ignore default value", func(t *testing.T) {
		mockService.Timeout = time.Second * 10
		t.Cleanup(func() {
			mockService.Timeout = time.Duration(0)
		})
		expected := mockService.GetTimeout()
		if defaultTimeout == expected {
			t.Errorf("timeout must be %s got %s", expected, mockService.Timeout)
		}

	})
}

func TestServiceConfigurableTypeNil(t *testing.T) {
	var serviceConfigurableTests = []struct {
		Name   string
		want   bool
		result any
		out    string
	}{
		{
			Name:   "rate limit must be nil",
			want:   mockService.RateLimitsIsNil(),
			result: mockService.RateLimits,
			out:    "rate limits expected to be nil, got %v",
		},
		{
			Name:   "circuit breaker must be nil",
			want:   mockService.CbIsNil(),
			result: mockService.CircuitBreaker,
			out:    "circuit breaker expected to be nil, got %v",
		},
		{
			Name:   "retry must be nil",
			want:   mockService.RetryIsNil(),
			result: mockService.RetryBackoff,
			out:    "retry expected to be nil, got %v",
		},
	}
	for _, test := range serviceConfigurableTests {
		t.Run(test.Name, func(t *testing.T) {
			if !test.want {
				t.Errorf(test.out, test.result)
			}
		})
	}
}

func TestServiceConfigurableTypeNotNil(t *testing.T) {
	service := mockServiceWithConfigurableTyple()
	//fmt.Println("Debug", !service.RateLimitsIsNil())
	var serviceConfigurableTests = []struct {
		Name   string
		want   bool
		result any
		out    string
	}{
		{
			Name:   "rate limit must not be nil",
			want:   service.RateLimitsIsNil(),
			result: service.RateLimits,
			out:    "rate limits expected to not be nil, got %v",
		},
		{
			Name:   "circuit breaker must not be nil",
			want:   service.CbIsNil(),
			result: service.CircuitBreaker,
			out:    "circuit breaker expected to not be nil, got %v",
		},
		{
			Name:   "retry must not be nil",
			want:   service.RetryIsNil(),
			result: service.RetryBackoff,
			out:    "retry expected to not be nil, got %v",
		},
	}
	for _, test := range serviceConfigurableTests {
		t.Run(test.Name, func(t *testing.T) {
			if test.want {
				t.Errorf(test.out, test.result)
			}
		})
	}
}
