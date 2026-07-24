package internal

import (
	"context"
	"fmt"
	netHttp "net/http"
	"time"

	"github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	healthHttp "github.com/AppsFlyer/go-sundheit/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// RegisterHealthChecks creates and registers a MongoDB health check.
// It returns an http.HandlerFunc that serves the health status in JSON,
// and a closer that disconnects the health-check MongoDB client on shutdown.
func RegisterHealthChecks(mongoURI string) (netHttp.HandlerFunc, func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	h := gosundheit.New()

	mongoCheck := &checks.CustomCheck{
		CheckName: "mongodb.ping",
		CheckFunc: func(ctx context.Context) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			return nil, client.Ping(ctx, readpref.Primary())
		},
	}

	err = h.RegisterCheck(mongoCheck,
		gosundheit.ExecutionPeriod(10*time.Second),
		gosundheit.InitialDelay(1*time.Second),
	)
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("failed to register MongoDB health check: %w", err)
	}

	closer := func(ctx context.Context) error {
		return client.Disconnect(ctx)
	}

	return healthHttp.HandleHealthJSON(h), closer, nil
}
