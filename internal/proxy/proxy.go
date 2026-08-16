package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
	"github.com/serge1997/apigateway/shared"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type ServiceHttpHandler func(http.HandlerFunc) http.HandlerFunc
type MiddlewareHandler func(ctx *Context, next http.HandlerFunc) (http.HandlerFunc, error)
type MiddlewareHandlerMap map[string]MiddlewareHandler
type CombinedMiddlewares = map[string]MiddlewareHandlerMap

var middlewares map[string]ServiceHttpHandler = map[string]ServiceHttpHandler{}

type Proxy struct {
	service     *service.Service
	Ctx         *Context
	middlewares CombinedMiddlewares
}

func New(service *service.Service, middlewares CombinedMiddlewares, w http.ResponseWriter, r *http.Request) *Proxy {
	return &Proxy{service: service, Ctx: &Context{r, w}, middlewares: middlewares}
}

func (p *Proxy) FullDomainePath() string {
	path := p.Ctx.Req().URL.Path
	query := p.Ctx.Req().URL.RawQuery
	return fmt.Sprintf("%s%s?%s", p.service.Target, path, query)
}
func (p *Proxy) Jwt() (string, error) {
	fullToken := p.Ctx.Req().Header.Get("Authorization")
	if fullToken != "" {
		fullTokenSplited := strings.Split(fullToken, " ")
		if len(fullTokenSplited) < 2 {
			return "", shared.ErrInvalidJWT
		}
		return fullTokenSplited[1], nil
	}
	return "", nil
}

func (p *Proxy) setDefaultHeaders(client *http.Request) {
	jwt := p.Ctx.Req().Header.Get("Authorization")
	contentType := p.Ctx.Req().Header.Get("Content-Type")
	client.Header.Set("Content-Type", contentType)
	if jwt != "" {
		client.Header.Set("Autorization", jwt)
	}
}

func (p *Proxy) splitMiddlewares() (before, after []MiddlewareHandler) {
	serviceMiddlewares := p.middlewares["service"]
	globalMdlwsName := slices.Collect(maps.Keys(p.middlewares["global"]))
	globalMdlwsHandler := slices.Collect(maps.Values(p.middlewares["global"]))
	after = append(after, globalMdlwsHandler...)
	for _, name := range p.service.Middlewares {
		handler, ok := serviceMiddlewares[name]
		if !ok {
			continue
		}
		if slices.Contains(globalMdlwsName, name) {
			continue
		}
		if slices.Contains(p.service.CbConfig.Before, name) {
			before = append(before, handler)
		} else {
			after = append(after, handler)
		}
	}
	return
}
func (p *Proxy) buildMiddlewaresChain(handler http.HandlerFunc, afterMiddlewares []MiddlewareHandler) http.HandlerFunc {
	finalHandler := handler
	for _, middleware := range afterMiddlewares {
		handler_, err := middleware(p.Ctx, finalHandler)
		if err != nil {
			http.Error(p.Ctx.Writer(), p.Ctx.Err(err, 501).Error(), 501)
			return nil
		}
		finalHandler = handler_
	}
	return finalHandler
}
func (p *Proxy) applyRateLimitChain(handler http.HandlerFunc) http.HandlerFunc {
	serviceRateLimits := p.service.RateLimits
	if len(serviceRateLimits) == 0 {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for _, rateLimitConfig := range serviceRateLimits {
			token, _ := p.Jwt()                     //return empty jwt err in rateLimiter.Allow()
			id, _ := shared.ExtractIDFromJWT(token) // return empty jwt err in rateLimiter.Allow()
			pathToCamelcase := strings.ReplaceAll(p.Ctx.Path(), "/", "_")
			rateLimitConfig.Key = fmt.Sprintf("%s_%s%s_%s", p.service.Name, p.Ctx.Method(), pathToCamelcase, id)
			rateLimiter := ratelimit.NewRateLimiter(rateLimitConfig)
			if err := rateLimiter.Allow(); err != nil {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
		}
		handler(w, r)
	}
}
func (p *Proxy) withCircuitBreaker(handler http.HandlerFunc, befores []MiddlewareHandler) http.HandlerFunc {
	finalHandler := handler
	if p.service.CbConfig == nil {
		return handler
	}
	if len(befores) > 0 {
		for _, beforeHandler := range befores {
			hander_, err := beforeHandler(p.Ctx, finalHandler)
			if err != nil {
				http.Error(p.Ctx.Writer(), p.Ctx.Err(err, 501).Error(), 501)
				return nil
			}
			finalHandler = hander_
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if p.service.CircuitBreaker == nil {
			cb, err := circuitbreaker.New(p.service.CbConfig)
			if err != nil {
				http.Error(w, p.Ctx.Err(err, http.StatusInternalServerError).Error(), http.StatusInternalServerError)
				return
			}
			p.service.CircuitBreaker = cb
		}
		if err := p.service.CircuitBreaker.Handle(); err != nil {
			http.Error(w, p.Ctx.Err(err, http.StatusBadGateway).Error(), http.StatusBadGateway)
			return
		}
		finalHandler(w, r)
	}
}
func (p *Proxy) makeServiceCall() result.Result[httpresponse.HttpResponse] {
	ctx, cancel := context.WithTimeout(p.Ctx.req.Context(), p.service.GetTimeout())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, p.Ctx.Method(), p.FullDomainePath(), p.Ctx.Body())
	if err != nil {
		return result.Fail(httpresponse.FailResponse(err, http.StatusInternalServerError))
	}
	p.setDefaultHeaders(request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		var responseErr result.Result[httpresponse.HttpResponse]
		if errors.Is(err, ctx.Err()) {
			responseErr = result.Fail(httpresponse.FailResponse(err, http.StatusGatewayTimeout))
		} else {
			responseErr = result.Fail(httpresponse.FailResponse(err, http.StatusServiceUnavailable))
		}
		return responseErr
	}
	defer response.Body.Close()
	if response.StatusCode == 503 && p.service.CircuitBreaker != nil {
		p.service.CircuitBreaker.RecordFailure()
	}
	var clientResponse shared.HttpResponse
	if err = json.NewDecoder(response.Body).Decode(&clientResponse); err != nil {
		errMesage := fmt.Errorf("erro on decode service response. detail: %v", err)
		return result.Fail(httpresponse.FailResponse(errMesage, 501))
	}
	if response.StatusCode > 299 {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("%s", clientResponse.Message), 501))
	}
	if p.service.CircuitBreaker != nil {
		p.service.CircuitBreaker.RecordSuccess()
	}
	return result.Ok(httpresponse.SuccessResponse(clientResponse.Data, http.StatusOK, clientResponse.Message))
}
func (p *Proxy) call(w http.ResponseWriter, r *http.Request) {
	if !p.service.HasRetryBackoffConfigured() {
		serviceResult := p.makeServiceCall()
		if !serviceResult.IsSuccess() {
			http.Error(w, serviceResult.Value().Message, serviceResult.Value().Status)
			return
		}
		fmt.Fprint(w, serviceResult.Value().Json())
		return
	}
	retryBackoff := retry.New(p.service.RetryBackoff)
	retryBackoffResult := retryBackoff.Execute(p.makeServiceCall, p.service.CircuitBreaker)
	if !retryBackoffResult.IsSuccess() {
		http.Error(w, retryBackoffResult.Value().Message, 501)
		return
	}
	fmt.Fprint(w, retryBackoffResult.Value().Json())
}

func (p *Proxy) Call(ctx context.Context) {
	before, after := p.splitMiddlewares()
	handler := p.withCircuitBreaker(p.call, before)
	handler = p.applyRateLimitChain(handler)
	handler = p.buildMiddlewaresChain(handler, after)
	if handler == nil {
		return
	}
	handler(p.Ctx.Writer(), p.Ctx.Req())
}

func Middlewares() map[string]ServiceHttpHandler {
	return middlewares
}
