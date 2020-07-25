package inner

import (
	"context"
	"fmt"
	"github.com/ppussar/mongodb_exporter/inner/wrapper"
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
	mc, err := mongo.NewClient(options.Client().ApplyURI(uri))
	client := Connection{
		client:  mc,
	}
	return client, err
}

func (con Connection) Aggregate(db string, collection string, command string, ctx context.Context) (wrapper.ICursor, error) {
	var pipeline interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &pipeline)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	return con.client.Database(db).Collection(collection).Aggregate(ctx, pipeline, opts)
}

func (con Connection) Find(db string, collection string, command string, ctx context.Context) (wrapper.ICursor, error) {
	var bdoc interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &bdoc)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	return con.client.Database(db).Collection(collection).Find(ctx, &bdoc)
}
