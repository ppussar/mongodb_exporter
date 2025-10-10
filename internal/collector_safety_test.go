package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

func TestExtractMetricValue(t *testing.T) {
	collector := &Collector{
		config: Metric{MetricsAttribute: "count"},
	}

	tests := []struct {
		name     string
		result   bson.M
		expected float64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "float64 value",
			result:   bson.M{"count": float64(42.5)},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "int32 value",
			result:   bson.M{"count": int32(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			result:   bson.M{"count": int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "int value",
			result:   bson.M{"count": int(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:    "missing attribute",
			result:  bson.M{"other": 42},
			wantErr: true,
			errMsg:  "metric attribute 'count' not found",
		},
		{
			name:    "unsupported type",
			result:  bson.M{"count": "string"},
			wantErr: true,
			errMsg:  "unsupported metric value type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := collector.extractMetricValue(tt.result)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}

func TestExtractVarTagsValues(t *testing.T) {
	collector := &Collector{
		varTagValueNames: []string{"type", "status"},
	}

	tests := []struct {
		name     string
		result   bson.M
		expected []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "string values",
			result:   bson.M{"type": "apple", "status": "fresh"},
			expected: []string{"apple", "fresh"},
			wantErr:  false,
		},
		{
			name:     "mixed types",
			result:   bson.M{"type": "apple", "status": int32(1)},
			expected: []string{"apple", "1"},
			wantErr:  false,
		},
		{
			name:     "numeric values",
			result:   bson.M{"type": int64(123), "status": float64(45.6)},
			expected: []string{"123", "45.6"},
			wantErr:  false,
		},
		{
			name:    "missing tag attribute",
			result:  bson.M{"type": "apple"},
			wantErr: true,
			errMsg:  "tag attribute 'status' not found",
		},
		{
			name:    "unsupported type",
			result:  bson.M{"type": "apple", "status": []string{"array"}},
			wantErr: true,
			errMsg:  "unsupported tag value type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := collector.extractVarTagsValues(tt.result)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, values)
			}
		})
	}
}

func TestUpdateConnection(t *testing.T) {
	collector := &Collector{}
	
	// Test thread-safe connection update with nil (valid for this test)
	collector.UpdateConnection(nil)
	
	collector.mu.RLock()
	assert.Nil(t, collector.mongo)
	collector.mu.RUnlock()
}
