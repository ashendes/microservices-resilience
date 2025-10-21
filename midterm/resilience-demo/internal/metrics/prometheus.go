package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal tracks total HTTP requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	// RequestDuration tracks HTTP request duration
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "endpoint"},
	)

	// CircuitBreakerState tracks circuit breaker state (0=closed, 1=open, 2=half-open)
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service", "circuit_name"},
	)

	// CircuitBreakerFailures tracks circuit breaker failures
	CircuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"service", "circuit_name"},
	)

	// BulkheadActiveRequests tracks active requests in bulkhead
	BulkheadActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_active_requests",
			Help: "Number of active requests in bulkhead",
		},
		[]string{"service", "bulkhead_name"},
	)

	// BulkheadRejectedRequests tracks rejected requests by bulkhead
	BulkheadRejectedRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_rejected_requests_total",
			Help: "Total number of rejected requests by bulkhead",
		},
		[]string{"service", "bulkhead_name"},
	)

	// OrdersTotal tracks total orders by status
	OrdersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total number of orders",
		},
		[]string{"status"},
	)

	// InventoryLevel tracks current inventory level
	InventoryLevel = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "inventory_level",
			Help: "Current inventory level",
		},
		[]string{"item_id"},
	)

	// PaymentAmount tracks payment amounts
	PaymentAmount = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "payment_amount_dollars",
			Help:    "Payment amounts in dollars",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000},
		},
	)

	// ChaosFailureRate tracks chaos engineering failure simulations
	ChaosFailureRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "chaos_failure_enabled",
			Help: "Whether chaos failure mode is enabled (1=enabled, 0=disabled)",
		},
		[]string{"service"},
	)

	// ChaosSlowMode tracks slow response simulation
	ChaosSlowMode = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "chaos_slow_mode_enabled",
			Help: "Whether chaos slow mode is enabled (1=enabled, 0=disabled)",
		},
		[]string{"service"},
	)
)

// PrometheusMiddleware creates a Gin middleware for automatic metrics collection
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		RequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		RequestDuration.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}
