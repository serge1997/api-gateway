package circuitbreaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrUnacessibleService error = errors.New("o serviço está inacessivel no momento, tenta mais tarde")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	FailureThreshold int           `json:"-"`            //limit of failures before opening the circuit
	RetryTimeout     time.Duration `json:"-"`            //time in seconds to wait before retrying after the circuit is opened
	FailureCount     int           `json:"failureCount"` //current number of consecutive failures
	LastFailure      time.Time     `json:"lastFailure"`
	SuccessCount     int           `json:"successCount"`
	State            State         `json:"state"` //current state of the circuit (closed, open, half-open)
	Before           []string
	mu               sync.RWMutex
}

func New(config *Config) (*CircuitBreaker, error) {
	timeout, err := time.ParseDuration(config.RetryTimeout)
	if err != nil {
		return nil, fmt.Errorf("retry_timeout inválido %q: %w", config.RetryTimeout, err)
	}
	return &CircuitBreaker{
		FailureThreshold: config.FailureThreshold,
		RetryTimeout:     timeout,
		FailureCount:     0,
		SuccessCount:     0,
		State:            Closed,
		Before:           config.Before,
	}, nil
}
func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailureCount++
	c.SuccessCount = 0
	c.LastFailure = time.Now()
	if c.FailureCount >= c.FailureThreshold {
		c.setOpen()
	}
}
func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isHalfOpen() {
		c.setClosed()
		c.FailureCount = 0
		c.SuccessCount = 0
	}
	c.SuccessCount++
}

func (c *CircuitBreaker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setClosed()
	c.FailureCount = 0
	c.SuccessCount = 0
}

// nao usar lock aqui
// já chamado no Handler com lock
func (c *CircuitBreaker) recordHalfOpen() {
	if c.FailureCount >= 1 {
		if c.LastFailure.Add(c.RetryTimeout).Before(time.Now()) {
			c.setHalfOpen()
		}
	}
}

func (c *CircuitBreaker) IsOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State == Open
}
func (c *CircuitBreaker) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State == Closed
}
func (c *CircuitBreaker) IsHalfOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State == HalfOpen
}

func (c *CircuitBreaker) isOpen() bool {
	return c.State == Open
}

func (c *CircuitBreaker) isHalfOpen() bool {
	return c.State == HalfOpen
}

func (c *CircuitBreaker) setOpen() {
	c.State = Open
}
func (c *CircuitBreaker) setClosed() {
	c.State = Closed
}

func (c *CircuitBreaker) setHalfOpen() {
	c.State = HalfOpen
}

func (c *CircuitBreaker) Handle() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isOpen() {
		c.recordHalfOpen()
		if !c.isHalfOpen() {
			return ErrUnacessibleService
		}
	}
	return nil
}
