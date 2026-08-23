package service

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/goccy/go-yaml"
	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/internal/retry"
)

var serviceHeaderName string = "x-service-name"
var defaultTimeout = time.Second * 5
var services map[string]*Service

func init() {

}

type Service struct {
	Name           string                         `yaml:"name" json:"name"`
	Target         string                         `yaml:"target" json:"target"`
	Middlewares    []string                       `yaml:"middlewares" json:"middlewares"`
	RateLimits     []*ratelimit.Config            `yaml:"rate_limits" json:"rateLimits"`
	CircuitBreaker *circuitbreaker.CircuitBreaker `yam:"-" json:"cb"`
	RateLimiters   []ratelimit.RateLimiter        `yaml:"-" json:"rate_limiters"`
	CbConfig       *circuitbreaker.Config         `yaml:"circuit_breaker" json:"cbConfig"`
	Timeout        time.Duration                  `yaml:"timeout" json:"timeout"`
	RetryBackoff   *retry.RetryBackoffConfig      `yaml:"retry" json:"retryBackoff"`
}

func (s *Service) GetTimeout() time.Duration {
	if s.Timeout.Seconds() < float64(time.Second) {
		return defaultTimeout
	}
	return s.Timeout
}

type Servicess struct {
	Services map[string]Service
}

func Parse(data []byte) error {
	var config struct {
		Services map[string]*Service `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	services = config.Services
	for serviceName, service := range services {
		for _, before := range service.CbConfig.Before {
			if !slices.Contains(service.Middlewares, before) {
				return fmt.Errorf("middleware %q declarado no before não existe em middlewares no serviço %q", before, serviceName)
			}
		}
	}
	return nil
}

func Get(xServiceName string) *Service {
	val, ok := services[xServiceName]
	if !ok {
		return nil
	}
	return val
}

func GetFromRequest(r *http.Request) *Service {
	xServiceName := r.Header.Get(serviceHeaderName)
	return Get(xServiceName)
}

func Services() map[string]*Service {
	return services
}

func Timeout() time.Duration {
	return defaultTimeout
}

func (s *Service) HasRetryBackoffConfigured() bool {
	if s.RetryBackoff == nil {
		return false
	}
	return true
}

func (s *Service) CbIsNil() bool {
	return s.CbConfig == nil
}

func (s *Service) RateLimitsIsNil() bool {
	return s.RateLimits == nil
}

func (s *Service) RetryIsNil() bool {
	return s.RetryBackoff == nil
}

func (s *Service) Cb() *circuitbreaker.CircuitBreaker {
	return s.CircuitBreaker
}

func (s *Service) RtLimiters() []ratelimit.RateLimiter {
	return s.RateLimiters
}

func (s *Service) Allow() error {
	for _, limiter := range s.RateLimiters {
		if err := limiter.Allow(); err != nil {
			return err
		}
	}
	return nil
}
