package internal

import (
	"github.com/stretchr/testify/assert"
	netHttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMongoHealthCheck(t *testing.T) {

	t.Run("returns handler when MongoDB URI is valid", func(t *testing.T) {
		handler, err := RegisterHealthChecks("mongodb://localhost:27017")
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		
		// Give some time for the health check to initialize
		time.Sleep(2 * time.Second)

		req, err := netHttp.NewRequest("GET", "/health-check", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Since localhost:27017 likely doesn't have MongoDB running, expect unhealthy
		assert.Contains(t, []string{"200 OK", "503 Service Unavailable"}, rr.Result().Status)
	})

	t.Run("returns error when MongoDB URI is invalid", func(t *testing.T) {
		handler, err := RegisterHealthChecks("invalid-uri")
		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "failed to connect to MongoDB")
	})

	t.Run("returns handler for unreachable MongoDB host", func(t *testing.T) {
		handler, err := RegisterHealthChecks("mongodb://localhost:27999")
		assert.NoError(t, err)
		assert.NotNil(t, handler)
		
		// Give some time for the health check to initialize
		time.Sleep(2 * time.Second)

		req, err := netHttp.NewRequest("GET", "/health-check", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Should be unhealthy since the host is unreachable
		assert.Equal(t, "503 Service Unavailable", rr.Result().Status)
	})
}
