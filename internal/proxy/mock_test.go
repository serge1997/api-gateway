package proxy

import (
	"log"
	"net/http"
	"net/http/httptest"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/contracts"
	"github.com/serge1997/apigateway/internal/service"
)

var middlewaresMock MiddlewareHandlerMap = MiddlewareHandlerMap{
	"logger": func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("%s - %s", ctx.Method(), ctx.Path())
		return next, nil
	},
	"auth": func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		if ctx.Header("Authorization") == "" {
			return nil, ctx.Unauthorized()
		}
		return next, nil
	},
	"permission": func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		return next, nil
	},
}

var reqMock = func(method string, url string) *http.Request {
	return httptest.NewRequest(method, url, nil)
}

var writerMock = httptest.NewRecorder()
var serviceMock = &service.Service{
	Name:        "posts",
	Target:      "http://localhost",
	Middlewares: []string{"auth", "logger", "permission"},
	CbConfig: &circuitbreaker.Config{
		FailureThreshold: 5,
		RetryTimeout:     "500ms",
		Before:           []string{"logger"},
	},
}
var proxyMock = New(serviceMock, middlewaresMock, writerMock, reqMock(http.MethodGet, serviceMock.Target))
