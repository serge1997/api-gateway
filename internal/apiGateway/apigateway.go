package apigateway

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	"github.com/serge1997/apigateway/internal/middleware"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
	"github.com/serge1997/apigateway/internal/service"
)

var services = map[string]*service.Service{}

type apiGateway struct {
	Server   APIGatewayConfig            `yaml:"server" json:"server"`
	Services map[string]*service.Service `yaml:"services" json:"services"`
	//apply retry globaly(for all services)
	Retry          *retry.RetryBackoffConfig      `yaml:"retry" json:"retry"`
	CircuitBreaker *circuitbreaker.CircuitBreaker `yaml:"-"`
	//apply circuit breaker globaly(for all services)
	CircuitbreakerConfig *circuitbreaker.Config `yaml:"circuit_breaker" json:"circuit_breaker"`
	//apply rate limit globaly(for all services)
	RateLimits []*ratelimit.Config `yaml:"rate_limits" json:"rate_limits"`
	//Global middleware for all services requests
	Middlewares []string
	//Custom for extract service name from client request
	ServiceLookupHeader string `yaml:"service_lookup_header" json:"service_lookup_header"`
}
type DefaultOpts func(a *apiGateway) error

func New() *apiGateway {
	gtw, err := parse()
	if err != nil || gtw == nil {
		panic(err)
	}

	err = gtw.withDefaultOptions(
		applyGlobalDefaults(),
		applyDefaultConfigDefaultValue(),
	)
	if err != nil {
		panic(err)
	}
	cORSAllowedOrigins = gtw.Server.CORSAllowedOrigins
	defaultServiceLookupHeader = gtw.ServiceLookupHeader
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

func (a *apiGateway) Timeout() time.Duration {
	return a.Server.Timeout
}
func (a *apiGateway) ReadHeaderTimeout() time.Duration {
	return a.Server.ReadHeaderTimeout
}

func applyGlobalDefaults() DefaultOpts {
	return func(a *apiGateway) error {
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

			if len(service.Middlewares) == 0 {
				service.Middlewares = a.Middlewares
			} else {
				for _, name := range a.Middlewares {
					if !slices.Contains(service.Middlewares, name) {
						service.Middlewares = append(service.Middlewares, name)
					}
				}
			}

			if !service.CbIsNil() && len(service.CbConfig.Before) >= 1 {
				for _, before := range service.CbConfig.Before {
					if !slices.Contains(service.Middlewares, before) {
						return fmt.Errorf("You need to register a before [%s] middleware in middleware group", before)
					}
				}
			}
		}
		return nil
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

func Services() map[string]*service.Service {
	return services
}

func applyDefaultConfigDefaultValue() DefaultOpts {
	return func(a *apiGateway) error {
		a.serviceLookupHeader()
		return nil
	}
}

func (a *apiGateway) serviceLookupHeader() {
	if a.ServiceLookupHeader == "" {
		a.ServiceLookupHeader = defaultServiceLookupHeader
	}
}

func (a *apiGateway) withDefaultOptions(opts ...DefaultOpts) error {
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return err
		}
	}
	return nil
}

func ExtractServiceName(r *http.Request) string {
	return r.Header.Get(defaultServiceLookupHeader)
}
