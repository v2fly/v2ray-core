package deferredpersistentstorage

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/app/persistentstorage"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
)

type DeferredPersistentStorage interface {
	storage.ScopedPersistentStorage
	ProvideInner(ctx context.Context, inner persistentstorage.ScopedPersistentStorage)
}

var errNotExist = errors.New("persistent storage does not exist")

type deferredPersistentStorage struct {
	ready context.Context
	done  context.CancelFunc
	inner persistentstorage.ScopedPersistentStorage

	awaitingChildren []*deferredPersistentStorage

	intoScopes []string
}

func (d *deferredPersistentStorage) ScopedPersistentStorageEngine() {
}

func (d *deferredPersistentStorage) Put(ctx context.Context, key []byte, value []byte) error {
	<-d.ready.Done()
	if d.inner == nil {
		return errNotExist
	}
	return d.inner.Put(ctx, key, value)
}

func (d *deferredPersistentStorage) Get(ctx context.Context, key []byte) ([]byte, error) {
	<-d.ready.Done()
	if d.inner == nil {
		return nil, errNotExist
	}
	return d.inner.Get(ctx, key)
}

func (d *deferredPersistentStorage) List(ctx context.Context, keyPrefix []byte) ([][]byte, error) {
	<-d.ready.Done()
	if d.inner == nil {
		return nil, errNotExist
	}
	return d.inner.List(ctx, keyPrefix)
}

func (d *deferredPersistentStorage) Clear(ctx context.Context) {
	<-d.ready.Done()
	if d.inner == nil {
		return
	}
	d.inner.Clear(ctx)
}

func (d *deferredPersistentStorage) NarrowScope(ctx context.Context, key []byte) (storage.ScopedPersistentStorage, error) {
	if d.ready.Err() != nil {
		return d.inner.NarrowScope(ctx, key)
	}
	ready, done := context.WithCancel(ctx)
	swallowCopyScopes := d.intoScopes
	dps := &deferredPersistentStorage{
		ready:      ready,
		done:       done,
		inner:      nil,
		intoScopes: append(swallowCopyScopes, string(key)),
	}
	d.awaitingChildren = append(d.awaitingChildren, dps)
	return dps, nil
}

func (d *deferredPersistentStorage) DropScope(ctx context.Context, key []byte) error {
	<-d.ready.Done()
	if d.inner == nil {
		return errNotExist
	}
	return d.inner.DropScope(ctx, key)
}

func (d *deferredPersistentStorage) ProvideInner(ctx context.Context, inner persistentstorage.ScopedPersistentStorage) {
	d.inner = inner
	if inner != nil {
		for _, scope := range d.intoScopes {
			newScope, err := inner.NarrowScope(ctx, []byte(scope))
			if err != nil {
				panic(err)
			}
			d.inner = newScope
		}
	}
	for _, child := range d.awaitingChildren {
		child.ProvideInner(ctx, d.inner)
	}
	d.done()
}

func NewDeferredPersistentStorage(ctx context.Context) DeferredPersistentStorage {
	ready, done := context.WithCancel(ctx)
	return &deferredPersistentStorage{
		ready: ready,
		done:  done,
		inner: nil,
	}
}
