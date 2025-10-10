package internal

import (
	"context"
	"fmt"
	httpClient "net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHttpServer(t *testing.T) {

	underTest := NewHttpServer(Config{
		HTTP: HTTP{
			Prometheus: "/metrics",
			Health:     "/health",
			Liveliness: "/live",
		},
		MongoDb: MongoDB{URI: "mongodb://localhost:27017"},
	})

	serverRunWg := &sync.WaitGroup{}
	serverRunWg.Add(1)
	underTest.Start(serverRunWg)
	time.Sleep(2 * time.Second)

	t.Cleanup(func() {
		assert.NoError(t, underTest.Shutdown(context.TODO()))
	})

	t.Run("serves metrics endpoint", func(t *testing.T) {
		resp, err := httpClient.Get(fmt.Sprintf("http://localhost:%v/metrics", underTest.Port))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "200 OK", resp.Status)
	})

	t.Run("serves liveliness endpoint", func(t *testing.T) {
		resp, err := httpClient.Get(fmt.Sprintf("http://localhost:%v/live", underTest.Port))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "204 No Content", resp.Status)
	})

	t.Run("serves health endpoint", func(t *testing.T) {
		resp, err := httpClient.Get(fmt.Sprintf("http://localhost:%v/health", underTest.Port))
		if err != nil {
			t.Fatal(err)
		}
		// Health endpoint should respond (either healthy or unhealthy)
		assert.Contains(t, []string{"200 OK", "503 Service Unavailable"}, resp.Status)
	})
}
