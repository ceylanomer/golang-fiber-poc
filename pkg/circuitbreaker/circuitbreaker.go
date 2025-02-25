package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CircuitBreakerConfig struct {
	// Name is the identifier for this circuit breaker instance
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through when the CircuitBreaker is half-open
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for the CircuitBreaker to clear the internal counts
	Interval time.Duration

	// Timeout is the period of the open state, after which the state of the CircuitBreaker becomes half-open
	Timeout time.Duration

	// RequestsVolumeThreshold is the minimum number of requests needed before the CircuitBreaker can start evaluating failures
	RequestsVolumeThreshold uint32

	// FailureThreshold is the failure rate threshold in percentage (0.0 - 1.0). When the failure rate exceeds this value, the CircuitBreaker trips
	FailureThreshold float64
}

// NewCircuitBreaker creates a new circuit breaker with the given name
func NewCircuitBreaker(config CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,

		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= config.RequestsVolumeThreshold &&
				failureRatio >= config.FailureThreshold
		},

		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			zap.L().Info("CircuitBreaker state changed", zap.String("name", name), zap.String("from", from.String()), zap.String("to", to.String()))
		},
	})
}
