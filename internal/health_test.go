package internal

import (
	"github.com/stretchr/testify/assert"
	netHttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMongoHealthCheck(t *testing.T) {

	t.Run("is healthy if given mongo host is reachable", func(t *testing.T) {
		handler, err := RegisterHealthChecks("http://github.com")
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)

		req, err := netHttp.NewRequest("GET", "/health-check", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, "200 OK", rr.Result().Status)
	})

	t.Run("is unhealthy if given mongo host is not reachable", func(t *testing.T) {
		handler, err := RegisterHealthChecks("http://localhost:0")
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)

		req, err := netHttp.NewRequest("GET", "/health-check", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, "503 Service Unavailable", rr.Result().Status)
	})
}
