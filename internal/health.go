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
// It returns an http.HandlerFunc that serves the health status in JSON.
func RegisterHealthChecks(mongoURI string) (netHttp.HandlerFunc, error) {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Create gosundheit instance
	h := gosundheit.New()

	// Create MongoDB ping check
	mongoCheck := &checks.CustomCheck{
		CheckName: "mongodb.ping",
		CheckFunc: func(ctx context.Context) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			return nil, client.Ping(ctx, readpref.Primary())
		},
	}

	// Register the MongoDB ping check
	err = h.RegisterCheck(mongoCheck,
		gosundheit.ExecutionPeriod(10*time.Second),
		gosundheit.InitialDelay(1*time.Second),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register MongoDB health check: %w", err)
	}

	return healthHttp.HandleHealthJSON(h), nil
}
