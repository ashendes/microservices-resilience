package patterns

import (
	"fmt"
	"time"

	"github.com/ashendes/resilience-demo/internal/metrics"
	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
)

// CircuitBreakerWrapper wraps gobreaker with metrics
type CircuitBreakerWrapper struct {
	*gobreaker.CircuitBreaker
	name    string
	service string
}

// NewCircuitBreaker creates a new circuit breaker with Prometheus metrics
func NewCircuitBreaker(name, service string) *CircuitBreakerWrapper {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,                // Max requests allowed in half-open state
		Interval:    15 * time.Second, // Window to track failures
		Timeout:     30 * time.Second, // Time to wait before half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			// Trip if 60% or more requests fail and at least 3 requests have been made
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(cbName string, from gobreaker.State, to gobreaker.State) {
			// Update Prometheus metrics
			state := float64(0)
			switch to {
			case gobreaker.StateOpen:
				state = 1
			case gobreaker.StateHalfOpen:
				state = 2
			case gobreaker.StateClosed:
				state = 0
			}
			metrics.CircuitBreakerState.WithLabelValues(service, cbName).Set(state)

			log.WithFields(log.Fields{
				"circuit": cbName,
				"from":    from.String(),
				"to":      to.String(),
			}).Info("Circuit breaker state changed")
		},
	})

	wrapper := &CircuitBreakerWrapper{
		CircuitBreaker: cb,
		name:           name,
		service:        service,
	}

	// Initialize the metric with the current state (closed by default)
	metrics.CircuitBreakerState.WithLabelValues(service, name).Set(0)

	return wrapper
}

// Execute runs a function through the circuit breaker with metrics
func (cb *CircuitBreakerWrapper) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := cb.CircuitBreaker.Execute(fn)

	if err != nil {
		metrics.CircuitBreakerFailures.WithLabelValues(cb.service, cb.name).Inc()
	}

	return result, err
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreakerWrapper) GetState() string {
	return cb.State().String()
}

// GetStateValue returns numeric value for the state (0=closed, 1=open, 2=half-open)
func (cb *CircuitBreakerWrapper) GetStateValue() int {
	switch cb.State() {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateOpen:
		return 1
	case gobreaker.StateHalfOpen:
		return 2
	default:
		return -1
	}
}

// FormatError formats an error message with circuit breaker info
func FormatError(circuitName string, err error) error {
	if err == gobreaker.ErrOpenState {
		return fmt.Errorf("circuit breaker %s is open (service unavailable)", circuitName)
	}
	if err == gobreaker.ErrTooManyRequests {
		return fmt.Errorf("circuit breaker %s: too many requests in half-open state", circuitName)
	}
	return err
}
