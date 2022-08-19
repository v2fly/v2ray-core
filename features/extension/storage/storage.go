package storage

import (
	"context"
)

type ScopedPersistentStorage interface {
	ScopedPersistentStorageEngine()
	Put(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte) ([]byte, error)
	List(ctx context.Context, keyPrefix []byte) ([][]byte, error)
	Clear(ctx context.Context)
	NarrowScope(ctx context.Context, key []byte) (ScopedPersistentStorage, error)
	DropScope(ctx context.Context, key []byte) error
}

type ScopedTransientStorage interface {
	ScopedTransientStorage()
	Put(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (interface{}, error)
	List(ctx context.Context, keyPrefix string) ([]string, error)
	Clear(ctx context.Context)
	NarrowScope(ctx context.Context, key string) (ScopedTransientStorage, error)
	DropScope(ctx context.Context, key string) error
}
