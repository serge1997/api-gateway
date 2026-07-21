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
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/database"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/reporitory"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
	"github.com/serge1997/apigateway/internal/stream"
	"github.com/serge1997/apigateway/shared"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
)

type ServiceHttpHandler func(http.HandlerFunc) http.HandlerFunc

var middlewares map[string]ServiceHttpHandler = map[string]ServiceHttpHandler{}
var mux = http.NewServeMux()

type proxy struct {
	service  *service.Service
	req      *http.Request
	w        http.ResponseWriter
	duration time.Duration
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

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		serviceCollection := slices.Collect(maps.Values(service.Services()))
		response := shared.HttpResponse{Data: serviceCollection, Message: "todos os serviços", Status: 200}.Json()
		fmt.Fprint(w, response)
	})
	mux.HandleFunc("/api/summaries", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		repo := reporitory.New(database.Db())
		latencies, err := repo.AvgLatency(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		errRates, err := repo.ErrRate(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reqInLastMinute, err := repo.ReqInMinute(ctx, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reqInLastTenMinutes, err := repo.ReqInMinute(ctx, 10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reqInCountByMinutes, err := repo.ReqInMinuteGroupedByMinute(ctx, 10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, httpresponse.SuccessResponse(map[string]interface{}{
			"latencies":                       latencies,
			"errRates":                        errRates,
			"reqInLastMin":                    reqInLastMinute,
			"reqInLastTenMin":                 reqInLastTenMinutes,
			"reqLastTenMinutesCountByMinutes": reqInCountByMinutes,
		}, http.StatusOK, "").Json())
		return
	})

	mux.HandleFunc("/api/metrics-of-service", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		serviceName := r.URL.Query().Get("service")
		repo := reporitory.New(database.Db())
		httpStatusMetrics, err := repo.HttpStatusMetricsOfService(ctx, serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpStatusMethodsMetrics, err := repo.HttpMethodsMetricsOfService(ctx, serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		requestLast24Hours, err := repo.ReqInMinuteGroupedByMinuteOfService(ctx, 60, serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reqLogs, err := repo.AllInLastHourOfService(ctx, serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		durationMetrics, err := repo.DurationMetricsOfService(ctx, serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, httpresponse.SuccessResponse(map[string]interface{}{
			"statusMetrics":       httpStatusMetrics,
			"methodsMetrics":      httpStatusMethodsMetrics,
			"requestLastRequests": requestLast24Hours,
			"logs":                reqLogs,
			"durationMetrics":     durationMetrics,
		}, http.StatusOK, "").Json())
		return
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
		if slices.Contains(p.service.CbConfig.Before, name) {
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
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				reqStream := stream.NewRequestStream(p, result.Fail(httpresponse.FailResponse(
					err,
					http.StatusTooManyRequests,
				)))
				stream.Produce(reqStream)
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
		if p.service.CircuitBreaker == nil {
			cb, err := circuitbreaker.New(p.service.CbConfig)
			if err != nil {
				http.Error(w, err.Error(), 501)
				return
			}
			p.service.CircuitBreaker = cb
		}
		if err := p.service.CircuitBreaker.Handle(); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			p.service.CircuitBreaker.RecordFailure()
			reqStream := stream.NewRequestStream(p, result.Fail(httpresponse.FailResponse(
				err,
				http.StatusBadGateway,
			)))
			stream.Produce(reqStream)
			return
		}
		finalHandler(w, r)
	}
}
func (p *proxy) makeServiceCall() result.Result[httpresponse.HttpResponse] {
	ctx, cancel := context.WithTimeout(p.req.Context(), p.service.GetTimeout())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, p.req.Method, p.FullDomainePath(), p.req.Body)
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
func (p *proxy) call(w http.ResponseWriter, r *http.Request) {
	if !p.service.HasRetryBackoffConfigured() {
		serviceResult := p.makeServiceCall()
		if !serviceResult.IsSuccess() {
			http.Error(w, serviceResult.Value().Message, serviceResult.Value().Status)
			reqStream := stream.NewRequestStream(p, serviceResult)
			stream.Produce(reqStream)
			return
		}
		fmt.Fprint(w, serviceResult.Value().Json())
		return
	}
	retryBackoff := retry.New(p.service.RetryBackoff)
	retryBackoffResult := retryBackoff.Execute(p.makeServiceCall, p.service.CircuitBreaker)
	if !retryBackoffResult.IsSuccess() {
		http.Error(w, retryBackoffResult.Value().Message, 501)
		reqStream := stream.NewRequestStream(p, retryBackoffResult)
		stream.Produce(reqStream)
		return
	}
	fmt.Fprint(w, retryBackoffResult.Value().Json())
	reqStream := stream.NewRequestStream(p, retryBackoffResult)
	stream.Produce(reqStream)
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

func (p *proxy) Method() string {
	return p.req.Method
}

func (p *proxy) Path() string {
	return p.req.URL.Path
}

func (p *proxy) ServiceName() string {
	return p.service.Name
}

func (p *proxy) StreamHeaders() map[string]string {
	return map[string]string{
		stream.HeaderContentType: p.req.Header.Get(stream.HeaderContentType),
		stream.HeaderUserAgent:   p.req.Header.Get(stream.HeaderUserAgent),
	}
}

func (p *proxy) HeadersToJson() string {
	data, _ := json.Marshal(p.StreamHeaders())
	return fmt.Sprintf("%s", data)
}

func (p *proxy) Duration() time.Duration {
	return p.duration
}

func (p *proxy) setDuration(start time.Time) {
	p.duration = time.Since(start)
}

func (p *proxy) Host() string {
	return p.req.Host
}
