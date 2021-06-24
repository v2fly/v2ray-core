package extension

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/features"
)

type PersistentStorageEngine interface {
	features.Feature

	PersistentStorageEngine()
	Put(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte) ([]byte, error)
	List(ctx context.Context, keyPrefix []byte) ([][]byte, error)
}

type ScopedPersistentStorage interface {
	ScopedPersistentStorageEngine()

	Put(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte) ([]byte, error)
	List(ctx context.Context, keyPrefix []byte) ([][]byte, error)

	ClearIfCharacteristicMismatch(ctx context.Context, characteristic []byte) error
	NarrowScope(ctx context.Context, key []byte) (ScopedPersistentStorage, error)
}

type ScopedTransientStorage interface {
	ScopedTransientStorage()
	Put(ctx context.Context, key []byte, value interface{}) error
	Get(ctx context.Context, key []byte) (interface{}, error)
	List(ctx context.Context, keyPrefix []byte) ([][]byte, error)
	Clear(ctx context.Context)
	NarrowScope(ctx context.Context, key []byte) (ScopedPersistentStorage, error)
}
