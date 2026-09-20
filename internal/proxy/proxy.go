package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/contracts"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
	"github.com/serge1997/apigateway/shared"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type ServiceHttpHandler func(http.HandlerFunc) http.HandlerFunc
type MiddlewareHandler func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error)
type MiddlewareHandlerMap map[string]MiddlewareHandler

type Proxy struct {
	service     *service.Service
	Ctx         *Context
	middlewares MiddlewareHandlerMap
	duration    time.Duration
}

func New(service *service.Service, middlewares MiddlewareHandlerMap, w http.ResponseWriter, r *http.Request) *Proxy {
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
		client.Header.Set("Authorization", jwt)
	}
}

func (p *Proxy) splitMiddlewares() (before, after []MiddlewareHandler) {
	for _, name := range p.service.Middlewares {
		handler, ok := p.middlewares[name]
		if !ok {
			continue
		}
		if !p.service.CbIsNil() && slices.Contains(p.service.CbConfig.Before, name) {
			before = append(before, handler)
			continue
		}
		after = append(after, handler)
	}
	return
}
func (p *Proxy) buildMiddlewaresChain(handler http.HandlerFunc, afterMiddlewares []MiddlewareHandler) http.HandlerFunc {
	finalHandler := handler
	for _, middleware := range afterMiddlewares {
		handler_, err := middleware(p.Ctx, finalHandler)
		if err != nil {
			statusCode := p.Ctx.GetStatusString(err.Error())
			http.Error(p.Ctx.Writer(), err.Error(), statusCode)
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
		// check if service rate Limiters has not instantiate yet
		if len(p.service.RateLimiters) < 1 {
			var limitersSlice []ratelimit.RateLimiter
			for _, rateLimitConfig := range serviceRateLimits {
				token, _ := p.Jwt()
				id, _ := shared.ExtractIDFromJWT(token)
				pathToCamelcase := strings.ReplaceAll(p.Ctx.Path(), "/", "_")
				rateLimitConfig.Key = fmt.Sprintf("%s_%s%s_%s", p.service.Name, p.Ctx.Method(), pathToCamelcase, id)
				rateLimiter := ratelimit.NewRateLimiter(rateLimitConfig)
				limitersSlice = append(limitersSlice, rateLimiter)
			}
			p.service.RateLimiters = limitersSlice
		}
		if err := p.service.Allow(); err != nil {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		handler(w, r)
	}
}
func (p *Proxy) withCircuitBreaker(handler http.HandlerFunc, befores []MiddlewareHandler) http.HandlerFunc {
	finalHandler := handler
	if p.service.CbConfig == nil {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if len(befores) > 0 {
			for _, beforeHandler := range befores {
				hander_, err := beforeHandler(p.Ctx, finalHandler)
				if err != nil {
					statusCode := p.Ctx.GetStatusString(err.Error())
					http.Error(w, err.Error(), statusCode)
					return
				}
				finalHandler = hander_
			}
		}
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
			p.service.CircuitBreaker.RecordFailure()
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
	reqStart := time.Now()
	response, err := http.DefaultClient.Do(request)
	p.setDuration(reqStart)
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
		return result.Fail(httpresponse.FailResponse(errMesage, http.StatusInternalServerError))
	}
	if response.StatusCode > 299 {
		return result.Fail(httpresponse.FailResponse(fmt.Errorf("%s", clientResponse.Message), http.StatusInternalServerError))
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
	retryBackoffResult := retryBackoff.Execute(p.makeServiceCall, p.service)
	if !retryBackoffResult.IsSuccess() {
		http.Error(w, retryBackoffResult.Value().Json(), p.Ctx.GetStatusString(retryBackoffResult.Value().Json()))
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
	return map[string]ServiceHttpHandler{}
}

func (p *Proxy) Method() string {
	return p.Ctx.Method()
}

func (p *Proxy) Path() string {
	return p.Ctx.Path()
}

func (p *Proxy) ServiceName() string {
	return p.service.Name
}

func (p *Proxy) StreamHeaders() map[string]string {
	return map[string]string{}
}

func (p *Proxy) Duration() time.Duration {
	return p.duration
}

func (p *Proxy) setDuration(start time.Time) {
	p.duration = time.Since(start)
}

func (p *Proxy) Host() string {
	return p.Ctx.Host()
}
