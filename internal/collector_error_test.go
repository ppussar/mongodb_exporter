package internal

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSendError(t *testing.T) {
	tests := []struct {
		name        string
		bufferSize  int
		errorCount  int
		expectBlock bool
	}{
		{
			name:        "single error fits in buffer",
			bufferSize:  5,
			errorCount:  1,
			expectBlock: false,
		},
		{
			name:        "multiple errors fit in buffer",
			bufferSize:  5,
			errorCount:  3,
			expectBlock: false,
		},
		{
			name:        "errors exceed buffer - should not block",
			bufferSize:  2,
			errorCount:  5,
			expectBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorC := make(chan error, tt.bufferSize)
			collector := &Collector{
				errorC: errorC,
			}

			// Send errors
			start := time.Now()
			for i := 0; i < tt.errorCount; i++ {
				collector.sendError(assert.AnError)
			}
			duration := time.Since(start)

			// Should never block for more than a few milliseconds
			assert.Less(t, duration, 100*time.Millisecond, "sendError should not block")

			// Check how many errors were actually sent
			close(errorC)
			receivedCount := 0
			for range errorC {
				receivedCount++
			}

			if tt.errorCount <= tt.bufferSize {
				assert.Equal(t, tt.errorCount, receivedCount, "All errors should be sent when buffer has space")
			} else {
				assert.LessOrEqual(t, receivedCount, tt.bufferSize, "Should not exceed buffer size")
			}
		})
	}
}

func TestCollectorNilConnection(t *testing.T) {
	errorC := make(chan error, 1)
	collector := &Collector{
		config: Metric{Name: "test_metric"},
		mongo:  nil,
		errorC: errorC,
	}

	// Create a channel to collect metrics
	metricC := make(chan prometheus.Metric, 1)
	
	// This should handle nil connection gracefully
	collector.Collect(metricC)

	// Should have sent an error
	select {
	case err := <-errorC:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no MongoDB connection available")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error for nil connection")
	}

	// Should not have sent any metrics
	select {
	case <-metricC:
		t.Fatal("Should not send metrics with nil connection")
	default:
		// Expected
	}
}
