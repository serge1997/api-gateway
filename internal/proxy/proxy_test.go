package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/serge1997/apigateway/internal/contracts"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
)

func TestProxyJwt(t *testing.T) {
	jwt, _ := proxyMock.Jwt()
	var z string
	t.Run("must be empty when no header", func(t *testing.T) {
		if jwt != z {
			t.Errorf("jwt must be empty, got %s", jwt)
		}
	})

	t.Run("must extract jwt token from header", func(t *testing.T) {
		expected := "jwt_mocked"
		proxyMock.Ctx.Req().Header.Set("Authorization", fmt.Sprintf("Bearer %s", expected))
		jwt, _ = proxyMock.Jwt()
		if jwt != expected {
			t.Errorf("jwt must be equal to %s, got %s", expected, jwt)
		}
	})
}

func TestProxyHasSetDefaultHeaders(t *testing.T) {
	client := httptest.NewRequest(http.MethodGet, "http://apitgateway", nil)
	proxyMock.Ctx.req.Header.Set("Content-Type", "multipart/form-data")
	proxyMock.Ctx.req.Header.Set("Authorization", "beaer_mocked_token")
	proxyMock.setDefaultHeaders(client)
	t.Run("proxy must set client content type header to service call request", func(t *testing.T) {
		expected := "multipart/form-data"
		if client.Header.Get("Content-Type") != expected {
			t.Errorf("content type header must be %s got %s", expected, client.Header.Get("Content-Type"))
		}
	})

	t.Run("proxy must set client Authorization header to service call request", func(t *testing.T) {
		expected := "beaer_mocked_token"
		if client.Header.Get("Authorization") != expected {
			t.Errorf("Authorization header must be %s got %s", expected, client.Header.Get("Authorization"))
		}
	})
}

func TestProxyMustSplitMiddleware(t *testing.T) {
	before, after := proxyMock.splitMiddlewares()
	t.Run("before must contain expected number of middlewares configured", func(t *testing.T) {
		expected := 1
		if len(before) != 1 {
			t.Errorf("before must contain %d middleware configured got %d", expected, len(before))
		}
	})

	t.Run("after must contain expected number of middlewares configured", func(t *testing.T) {
		expected := 2
		if len(after) != expected {
			t.Errorf("after must contain %d middlewares configured got %d", expected, len(after))
		}
	})
}

func TestBuildMiddlewaresChain(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final handler"))
	}
	t.Run("must execute final handler on nil afters middlewares", func(t *testing.T) {
		result := proxyMock.buildMiddlewaresChain(handler, nil)
		recorder := httptest.NewRecorder()
		result(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		if recorder.Body.String() != "final handler" {
			t.Errorf("expected final handler output")
		}
	})

	t.Run("must fail on after middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()
		proxyMock = New(serviceMock, middlewaresMock, recorder, req)
		_, after := proxyMock.splitMiddlewares()
		_ = proxyMock.buildMiddlewaresChain(handler, after)
		body := recorder.Body.String()
		var m map[string]interface{}
		json.Unmarshal([]byte(body), &m)
		success := m["success"].(bool)
		if success != false {
			t.Error("handler must fail on failed middleware")
		}
	})

	t.Run("middlewares chain must successed and execute final handler", func(t *testing.T) {
		middlewaresMock["auth"] = func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
			return next, nil
		}
		_, after := proxyMock.splitMiddlewares()
		recorder := httptest.NewRecorder()
		result := proxyMock.buildMiddlewaresChain(handler, after)
		result(recorder, httptest.NewRequest("GET", serviceMock.Target, nil))
		if recorder.Body.String() != "final handler" {
			t.Errorf("all middlewares must success and execute correctly a final handler")
		}
	})
}

func TestApplyRateLimitChain(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("rate limit final handler"))
	}
	req := httptest.NewRequest(http.MethodGet, serviceMock.Target, nil)

	t.Run("must return final handler on empty rate limit", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		resultHandler := proxyMock.applyRateLimitChain(handler)
		resultHandler(recorder, req)
		if recorder.Body.String() != "rate limit final handler" {
			t.Error("must return final handler")
		}
	})

	t.Run("must fail with rate limite error", func(t *testing.T) {
		t.Cleanup(func() {
			for _, rt := range serviceMock.RateLimiters {
				rt.ClearCacheForTest(t)
			}
		})
		serviceMock.RateLimits = []*ratelimit.Config{
			{
				Type:  "bucket",
				Rate:  2,
				Burst: 1,
			},
		}
		resultHandler := proxyMock.applyRateLimitChain(handler)
		var expected string = "bucket limit exceeded"
		var result string
		for range 3 {
			recorder := httptest.NewRecorder()
			resultHandler(recorder, req)
			result = strings.Trim(recorder.Body.String(), "\n")
		}
		if result != expected {
			t.Errorf("must fail with bucket rate limit erro: %s got %s", expected, result)
		}
	})

	t.Run("rate limit must allow final handler execution", func(t *testing.T) {
		serviceMock.RateLimiters = nil
		serviceMock.RateLimits = nil
		serviceMock.RateLimits = []*ratelimit.Config{
			{
				Type:  "bucket",
				Rate:  5,
				Burst: 10,
			},
		}

		resultHandler := proxyMock.applyRateLimitChain(handler)
		var expected string = "rate limit final handler"
		var result string
		for range 3 {
			recorder := httptest.NewRecorder()
			resultHandler(recorder, req)
			result = recorder.Body.String()
		}
		if result != expected {
			t.Errorf("must success with final handler response: %s got %s", expected, result)
		}
	})

}
