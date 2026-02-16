package patterns

import (
	"context"
	"time"
)

// WithTimeout creates a context with timeout for fail-fast behavior
func WithTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

// DefaultTimeout is the default timeout for HTTP requests
const DefaultTimeout = 3 * time.Second

// SlowServiceTimeout is a longer timeout for services that might be slow
const SlowServiceTimeout = 10 * time.Second
