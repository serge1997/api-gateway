package service

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/goccy/go-yaml"
	circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"
	ratelimit "github.com/serge1997/apigateway/internal/rateLimit"
	"github.com/serge1997/apigateway/shared"
)

var serviceHeaderName string = "x-service-name"
var services map[string]Service

func init() {
	services, err := shared.LoadServiceYml()
	if err != nil {
		panic(err)
	}

	if err := Parse(services); err != nil {
		panic(err)
	}
}

type Service struct {
	Name           string                         `yaml:"name"`
	Target         string                         `yaml:"target"`
	Middlewares    []string                       `yaml:"middlewares"`
	RateLimits     []*ratelimit.Config            `yaml:"rate_limits"`
	CircuitBreaker *circuitbreaker.CircuitBreaker `yam:"-"`
	CbConfig       *circuitbreaker.Config         `yaml:"circuit_breaker"`
}
type Servicess struct {
	Services map[string]Service
}

func Parse(data []byte) error {
	var config struct {
		Services map[string]Service `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	services = config.Services
	for serviceName, service := range services {
		for _, middlewareName := range service.Middlewares {
			if !slices.Contains(service.CbConfig.Before, middlewareName) {
				return fmt.Errorf("middleware %q declarado no before não existe em middlewares", serviceName)
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
	return &val
}

func GetFromRequest(r *http.Request) *Service {
	xServiceName := r.Header.Get(serviceHeaderName)
	return Get(xServiceName)
}

func Services() map[string]Service {
	return services
}
