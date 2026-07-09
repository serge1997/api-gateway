package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/service"
	"github.com/serge1997/apigateway/shared"
)

type ServiceHttpHandler func(http.HandlerFunc) http.HandlerFunc

var middlewares map[string]ServiceHttpHandler = map[string]ServiceHttpHandler{}
var mux = http.NewServeMux()

type proxy struct {
	service *service.Service
	req     *http.Request
	w       http.ResponseWriter
}

func Use(name string, middleware ServiceHttpHandler) {
	middlewares[name] = middleware
}
func New(service *service.Service, r *http.Request, w http.ResponseWriter) *proxy {
	return &proxy{service: service, req: r, w: w}
}

func Router() http.Handler {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serviceName := r.Header.Get("x-service-name")
		srvce := service.Get(serviceName)
		if srvce == nil {
			fmt.Fprint(w, "service not found")
			return
		}
		prxy := New(srvce, r, w)
		prxy.Call(r.Context())
	})
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := shared.HttpResponse{Data: service.Services(), Message: "todos os serviços", Status: 200}.Json()
		fmt.Fprint(w, response)
	})
	return mux
}

func (p *proxy) FullDomainePath() string {
	path := p.req.URL.Path
	query := p.req.URL.RawQuery
	return fmt.Sprintf("%s%s?%s", p.service.Target, path, query)
}
func (p *proxy) Jwt() (string, error) {
	fullToken := p.req.Header.Get("Authorization")
	if fullToken != "" {
		fullTokenSplited := strings.Split(fullToken, " ")
		if len(fullTokenSplited) < 2 {
			return "", shared.ErrInvalidJWT
		}
		return fullTokenSplited[1], nil
	}
	return "", nil
}

func (p *proxy) setDefaultHeaders(client *http.Request) {
	jwt := p.req.Header.Get("Authorization")
	contentType := p.req.Header.Get("Content-Type")
	client.Header.Set("Content-Type", contentType)
	if jwt != "" {
		client.Header.Set("Autorization", jwt)
	}
}

func (p *proxy) splitMiddlewares() (before, after []ServiceHttpHandler) {
	for _, name := range p.service.Middlewares {
		handler, ok := middlewares[name]
		if !ok {
			continue
		}
		if slices.Contains(p.service.CircuitBreaker.Before(), name) {
			before = append(before, handler)
		} else {
			after = append(after, handler)
		}
	}
	return
}
func (p *proxy) buildMiddlewaresChain(handler http.HandlerFunc, afterMiddlewares []ServiceHttpHandler) http.HandlerFunc {
	finalHandler := handler
	for _, middleware := range afterMiddlewares {
		finalHandler = middleware(finalHandler)
	}
	return finalHandler
}
func (p *proxy) applyRateLimitChain(handler http.HandlerFunc) http.HandlerFunc {
	serviceRateLimits := p.service.RateLimits
	if len(serviceRateLimits) == 0 {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for _, rateLimitConfig := range serviceRateLimits {
			token, _ := p.Jwt()                     //return empty jwt err in rateLimiter.Allow()
			id, _ := shared.ExtractIDFromJWT(token) // return empty jwt err in rateLimiter.Allow()
			pathToCamelcase := strings.ReplaceAll(p.req.URL.Path, "/", "_")
			rateLimitConfig.Key = fmt.Sprintf("%s_%s%s_%s", p.service.Name, p.req.Method, pathToCamelcase, id)
			rateLimiter := ratelimit.NewRateLimiter(rateLimitConfig)
			if err := rateLimiter.Allow(); err != nil {
				fmt.Println(err)
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
		}
		handler(w, r)
	}
}
func (p *proxy) withCircuitBreaker(handler http.HandlerFunc, befores []ServiceHttpHandler) http.HandlerFunc {
	finalHandler := handler
	if p.service.CbConfig == nil {
		return handler
	}
	if len(befores) > 0 {
		for _, beforeHandler := range befores {
			finalHandler = beforeHandler(finalHandler)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var cb *circuitbreaker.CircuitBreaker
		if p.service.CircuitBreaker == nil {
			cb, err := circuitbreaker.New(p.service.CbConfig)
			if err != nil {
				http.Error(w, err.Error(), 501)
				return
			}
			p.service.CircuitBreaker = cb
		}
		if err := cb.Handle(); err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
		finalHandler(w, r)
	}
}
func (p *proxy) call(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(p.req.Context(), p.service.GetTimeout())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, p.req.Method, p.FullDomainePath(), p.req.Body)
	if err != nil {
		responseErr := shared.HttpResponse{Message: err.Error(), Status: 501}.Json()
		http.Error(w, responseErr, http.StatusInternalServerError)
		return
	}
	p.setDefaultHeaders(request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			responseErr := shared.HttpResponse{Message: fmt.Sprintf("service call timeout. err: %s", err.Error()), Status: http.StatusGatewayTimeout}.Json()
			http.Error(w, responseErr, http.StatusGatewayTimeout)
			return
		}
		responseErr := shared.HttpResponse{Message: err.Error(), Status: 501}.Json()
		http.Error(w, responseErr, response.StatusCode)
		if p.service.CircuitBreaker != nil {
			p.service.CircuitBreaker.RecordFailure()
		}
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 503 && p.service.CircuitBreaker != nil {
		p.service.CircuitBreaker.RecordFailure()
	}
	var clientResponse shared.HttpResponse
	if err = json.NewDecoder(response.Body).Decode(&clientResponse); err != nil {
		errMesage := fmt.Errorf("erro on decode service response. detail: %v", err)
		response := shared.HttpResponse{Message: errMesage.Error(), Status: 501}.Json()
		http.Error(w, response, 501)
		return
	}
	if response.StatusCode > 299 {
		responseErr := shared.HttpResponse{Message: clientResponse.Message, Status: response.StatusCode}.Json()
		http.Error(w, responseErr, response.StatusCode)
		return
	}
	successResponse := shared.HttpResponse{Message: clientResponse.Message, Data: clientResponse.Data, Status: response.StatusCode}.Json()
	if p.service.CircuitBreaker != nil {
		p.service.CircuitBreaker.RecordSuccess()
	}
	fmt.Fprint(w, successResponse)
}
func (p *proxy) Call(ctx context.Context) {
	before, after := p.splitMiddlewares()
	handler := p.withCircuitBreaker(p.call, before)
	handler = p.applyRateLimitChain(handler)
	handler = p.buildMiddlewaresChain(handler, after)
	handler(p.w, p.req)
}

func Middlewares() map[string]ServiceHttpHandler {
	return middlewares
}
