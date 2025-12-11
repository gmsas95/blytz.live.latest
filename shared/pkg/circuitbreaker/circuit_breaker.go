package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	MaxRequests       uint32        // Maximum number of requests allowed when half-open
	Interval         time.Duration  // Time period for which to count failures
	Timeout          time.Duration  // Time to wait before switching from half-open to open
	ReadyToTrip     func(counts Counts) bool
	OnStateChange    func(name string, from State, to State)
	IsSuccessful    func(err error) bool
}

// Counts holds circuit breaker counters
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses  uint32
	ConsecutiveFailures   uint32
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(Counts) bool
	onStateChange func(string, State, State)
	isSuccessful  func(error) bool

	mutex      sync.Mutex
	state      State
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker creates a new circuit breaker with default configuration
func NewCircuitBreaker(name string) *CircuitBreaker {
	return NewCircuitBreakerWithConfig(name, DefaultConfig())
}

// NewCircuitBreakerWithConfig creates a new circuit breaker with custom configuration
func NewCircuitBreakerWithConfig(name string, config Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:        name,
		maxRequests: config.MaxRequests,
		interval:    config.Interval,
		timeout:     config.Timeout,
		readyToTrip: config.ReadyToTrip,
		onStateChange: func(name string, from State, to State) {
			if config.OnStateChange != nil {
				config.OnStateChange(name, from, to)
			}
		},
		isSuccessful: config.IsSuccessful,
	}

	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}

	if cb.interval <= 0 {
		cb.interval = time.Duration(0)
	}

	if cb.timeout <= 0 {
		cb.timeout = 60 * time.Second
	}

	if cb.readyToTrip == nil {
		cb.readyToTrip = DefaultReadyToTrip
	}

	if cb.isSuccessful == nil {
		cb.isSuccessful = DefaultIsSuccessful
	}

	cb.toNewGeneration(time.Now())

	return cb
}

// DefaultConfig returns a default configuration for circuit breaker
func DefaultConfig() Config {
	return Config{
		MaxRequests: 1,
		Interval:    0,
		Timeout:     60 * time.Second,
		ReadyToTrip: DefaultReadyToTrip,
		IsSuccessful: DefaultIsSuccessful,
	}
}

// DefaultReadyToTrip is the default trip condition
func DefaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

// DefaultIsSuccessful is the default success condition
func DefaultIsSuccessful(err error) bool {
	return err == nil
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer cb.afterRequest(generation, cb.isSuccessful(err))

	result, err := fn()
	return result, err
}

// ExecuteContext runs the given function with context if the circuit breaker allows it
func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer cb.afterRequest(generation, cb.isSuccessful(err))

	result, err := fn(ctx)
	return result, err
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current counters
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// beforeRequest is called before a request is executed
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrOpenState
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.Requests++
	return generation, nil
}

// afterRequest is called after a request is executed
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess is called when a request succeeds
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen {
		cb.setState(StateClosed, now)
	}
}

// onFailure is called when a request fails
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if state == StateClosed && cb.readyToTrip(cb.counts) {
		cb.setState(StateOpen, now)
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState sets the new state
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	cb.onStateChange(cb.name, prev, state)
}

// toNewGeneration creates a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// Errors
var (
	ErrOpenState       = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// HTTPErrorIsSuccessful checks if an HTTP error should be considered successful
func HTTPErrorIsSuccessful(err error) bool {
	if err == nil {
		return true
	}

	if httpErr, ok := err.(*HTTPError); ok {
		// Consider 4xx client errors as successful (circuit breaker shouldn't trip on client errors)
		// but 5xx server errors should trip the circuit breaker
		return httpErr.StatusCode < 500
	}

	return false
}