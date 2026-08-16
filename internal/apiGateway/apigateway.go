package apigateway

import (
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/middleware"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
)

type apiGateway struct {
	Server               APIGatewayConfig               `yaml:"server" json:"server"`
	Services             map[string]*service.Service    `yaml:"services" json:"services"`
	Retry                retry.RetryBackoffConfig       `yaml:"retry" json:"retry"`
	CircuitBreaker       *circuitbreaker.CircuitBreaker `yaml:"-"`
	CircuitbreakerConfig *circuitbreaker.Config         `yaml:"circuit_breaker" json:"circuit_breaker"`
	RateLimit            []ratelimit.Config             `yaml:"rate_limit" json:"rate_limits"`
}

func New(cfg Config) *apiGateway {
	gtw, err := parse()
	if err != nil || gtw == nil {
		panic(err)
	}
	cORSAllowedOrigins = gtw.Server.CORSAllowedOrigins
	return gtw
}

func (a *apiGateway) Use(name string, handler middleware.Handler) {
	middleware.Register(name, handler)
}

func (a *apiGateway) UseGlobal(name string, handler middleware.Handler) {
	middleware.RegisterGlobal(name, handler)
}

func CORSAllowedOrigins() []string {
	return cORSAllowedOrigins
}

func HasCORSAllowedOrigins() bool {
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
