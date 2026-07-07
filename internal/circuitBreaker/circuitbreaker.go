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
	failureThreshold int           //limit of failures before opening the circuit
	retryTimeout     time.Duration //time in seconds to wait before retrying after the circuit is opened
	failureCount     int           //current number of consecutive failures
	lastFailure      time.Time
	successCount     int
	state            State //current state of the circuit (closed, open, half-open)
	before           []string
	mu               sync.RWMutex
}

func New(config *Config) (*CircuitBreaker, error) {
	timeout, err := time.ParseDuration(config.RetryTimeout)
	if err != nil {
		return nil, fmt.Errorf("retry_timeout inválido %q: %w", config.RetryTimeout, err)
	}
	return &CircuitBreaker{
		failureThreshold: config.FailureThreshold,
		retryTimeout:     timeout,
		failureCount:     0,
		successCount:     0,
		state:            Closed,
		before:           config.Before,
	}, nil
}
func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failureCount++
	c.successCount = 0
	c.lastFailure = time.Now()
	if c.failureCount >= c.failureThreshold {
		c.setOpen()
	}
}
func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isHalfOpen() {
		c.setClosed()
		c.failureCount = 0
		c.successCount = 0
	}
	c.successCount++
}

func (c *CircuitBreaker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setClosed()
	c.failureCount = 0
	c.successCount = 0
}

// nao usar lock aqui
// já chamado no Handler com lock
func (c *CircuitBreaker) recordHalfOpen() {
	if c.failureCount >= 1 {
		if c.lastFailure.Add(c.retryTimeout).Before(time.Now()) {
			c.setHalfOpen()
		}
	}
}

func (c *CircuitBreaker) IsOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == Open
}
func (c *CircuitBreaker) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == Closed
}
func (c *CircuitBreaker) IsHalfOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == HalfOpen
}

func (c *CircuitBreaker) isOpen() bool {
	return c.state == Open
}

func (c *CircuitBreaker) isHalfOpen() bool {
	return c.state == HalfOpen
}

func (c *CircuitBreaker) setOpen() {
	c.state = Open
}
func (c *CircuitBreaker) setClosed() {
	c.state = Closed
}

func (c *CircuitBreaker) setHalfOpen() {
	c.state = HalfOpen
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

func (c *CircuitBreaker) Before() []string {
	return c.before
}
