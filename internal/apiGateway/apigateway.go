package apigateway

import (
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/middleware"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
)

var services = map[string]*service.Service{}

type apiGateway struct {
	Server               APIGatewayConfig               `yaml:"server" json:"server"`
	Services             map[string]*service.Service    `yaml:"services" json:"services"`
	Retry                *retry.RetryBackoffConfig      `yaml:"retry" json:"retry"`
	CircuitBreaker       *circuitbreaker.CircuitBreaker `yaml:"-"`
	CircuitbreakerConfig *circuitbreaker.Config         `yaml:"circuit_breaker" json:"circuit_breaker"`
	RateLimits           []*ratelimit.Config            `yaml:"rate_limits" json:"rate_limits"`
}

func New() *apiGateway {
	gtw, err := parse()
	if err != nil || gtw == nil {
		panic(err)
	}
	cORSAllowedOrigins = gtw.Server.CORSAllowedOrigins
	gtw.applyGlobalDefaults()
	services = gtw.Services
	return gtw
}

func (a *apiGateway) Use(name string, handler middleware.Handler) {
	middleware.Register(name, handler)
}

func (a *apiGateway) UseGlobal(name string, handler middleware.Handler) {
	middleware.RegisterGlobal(name, handler)
}

func Cors() []string {
	return cORSAllowedOrigins
}

func HasCORS() bool {
	return len(cORSAllowedOrigins) >= 1
}

func (a *apiGateway) ListenAddr() string {
	return a.Server.ListenAddr
}

func (a *apiGateway) ReadTimeout() time.Duration {
	return a.Server.ReadTimeout
}

func (a *apiGateway) WriteTimeout() time.Duration {
	return a.Server.WriteTimeout
}

func (a *apiGateway) Timeout() time.Duration {
	return a.Server.Timeout
}

func (a *apiGateway) applyGlobalDefaults() {
	for _, service := range a.Services {
		if a.CircuitbreakerConfig != nil && service.CbIsNil() {
			service.CbConfig = a.CircuitbreakerConfig
		}
		if len(a.RateLimits) >= 1 && service.RateLimitsIsNil() {
			service.RateLimits = a.RateLimits
		}
		if a.Retry != nil && service.RetryIsNil() {
			service.RetryBackoff = a.Retry
		}
	}
}

func GetService(xName string) *service.Service {
	val, ok := services[xName]
	if !ok {
		return nil
	}
	return val
}

func (a *apiGateway) CORSAllowedOrigins() []string {
	return a.Server.CORSAllowedOrigins
}
