package internal

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	httpClient "net/http"
	"sync"
	"testing"
	"time"
)

func TestHttpServer(t *testing.T) {

	underTest := NewHttpServer(Config{
		HTTP: http{
			Prometheus: "/metrics",
			Health:     "/health",
			Liveliness: "/live",
		},
		MongoDb: mongoDb{URI: "http://github.com"},
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
		assert.Equal(t, "200 OK", resp.Status)
	})
}
