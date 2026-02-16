package patterns

import (
	"fmt"
	"time"

	"github.com/ashendes/resilience-demo/internal/metrics"
)

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	semaphore chan struct{}
	name      string
	service   string
}

// NewBulkhead creates a new bulkhead with specified capacity
func NewBulkhead(size int, name, service string) *Bulkhead {
	return &Bulkhead{
		semaphore: make(chan struct{}, size),
		name:      name,
		service:   service,
	}
}

// Execute runs a function within the bulkhead's resource limits
func (b *Bulkhead) Execute(fn func() error) error {
	select {
	case b.semaphore <- struct{}{}:
		// Update active requests metric
		metrics.BulkheadActiveRequests.WithLabelValues(b.service, b.name).Inc()

		defer func() {
			<-b.semaphore
			metrics.BulkheadActiveRequests.WithLabelValues(b.service, b.name).Dec()
		}()

		return fn()

	case <-time.After(1 * time.Second):
		// Record rejection
		metrics.BulkheadRejectedRequests.WithLabelValues(b.service, b.name).Inc()
		return fmt.Errorf("bulkhead %s: timeout acquiring resource", b.name)
	}
}

// GetName returns the bulkhead name
func (b *Bulkhead) GetName() string {
	return b.name
}
