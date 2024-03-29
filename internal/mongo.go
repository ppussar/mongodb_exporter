package internal

import (
	"context"
	"fmt"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connection to mongoDB
type Connection struct {
	client  *mongo.Client
	Context context.Context
}

// NewConnection opens a connection to mongoDB by using the given uri
func NewConnection(uri string) (wrapper.IConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	client := Connection{
		client: mc,
	}

	return client, err
}

// Aggregate executes a given aggregate query on the mongodb
func (con Connection) Aggregate(ctx context.Context, db string, collection string, command string) (wrapper.ICursor, error) {
	var pipeline interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &pipeline)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	return con.client.Database(db).Collection(collection).Aggregate(ctx, pipeline, opts)
}

// Find executes a given find query on the mongodb
func (con Connection) Find(ctx context.Context, db string, collection string, command string) (wrapper.ICursor, error) {
	var bdoc interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &bdoc)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	return con.client.Database(db).Collection(collection).Find(ctx, &bdoc)
}
