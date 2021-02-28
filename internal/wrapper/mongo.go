package wrapper

import "context"

// IConnection interface of mongo.Database
type IConnection interface {
	Aggregate(ctx context.Context, db string, collection string, command string) (ICursor, error)
	Find(ctx context.Context, db string, collection string, command string) (ICursor, error)
}

// ICursor interface of mongo.Cursor
type ICursor interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Err() error
	Close(ctx context.Context) error
}
