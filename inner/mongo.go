package inner

import (
	"context"
	"fmt"
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
func NewConnection(uri string) (Connection, error) {
	mc, err := mongo.NewClient(options.Client().ApplyURI(uri))
	ctx := context.Background()
	mc.Connect(ctx)
	client := Connection{
		client:  mc,
		Context: ctx,
	}
	return client, err
}

func (con Connection) aggregate(db string, collection string, command string) (*mongo.Cursor, error) {
	var pipeline interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &pipeline)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	return con.client.Database(db).Collection(collection).Aggregate(con.Context, pipeline, opts)
}

func (con Connection) find(db string, collection string, command string) (*mongo.Cursor, error) {
	var bdoc interface{}
	err := bson.UnmarshalExtJSON([]byte(command), true, &bdoc)
	if err != nil {
		fmt.Println(command)
		return nil, err
	}
	return con.client.Database(db).Collection(collection).Find(con.Context, &bdoc)
}
