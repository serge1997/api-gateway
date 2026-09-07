package contracts

import circuitbreaker "github.com/serge1997/apigateway/internal/circuitBreaker"

type ServiceRetry interface {
	Cb() *circuitbreaker.CircuitBreaker
	CbIsNil() bool
	Allow() error
}
