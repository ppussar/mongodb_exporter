package wrapper

import "context"

type IConnection interface {
	Aggregate(db string, collection string, command string, ctx context.Context) (ICursor, error)
	Find(db string, collection string, command string, ctx context.Context) (ICursor, error)
}

type ICursor interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Err() error
	Close(ctx context.Context) error
}
