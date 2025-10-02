package transientstorageimpl

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"strings"
	"sync"

	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
)

func NewScopedTransientStorageImpl() storage.ScopedTransientStorage {
	return &scopedTransientStorageImpl{scopes: map[string]storage.ScopedTransientStorage{}, values: map[string]interface{}{}}
}

type scopedTransientStorageImpl struct {
	access sync.Mutex
	scopes map[string]storage.ScopedTransientStorage
	values map[string]interface{}
}

func (s *scopedTransientStorageImpl) ScopedTransientStorage() {
	panic("implement me")
}

func (s *scopedTransientStorageImpl) Put(ctx context.Context, key string, value interface{}) error {
	s.access.Lock()
	defer s.access.Unlock()
	s.values[key] = value
	return nil
}

func (s *scopedTransientStorageImpl) Get(ctx context.Context, key string) (interface{}, error) {
	s.access.Lock()
	defer s.access.Unlock()
	sw, ok := s.values[key]
	if !ok {
		return nil, newError("unable to find ")
	}
	return sw, nil
}

func (s *scopedTransientStorageImpl) List(ctx context.Context, keyPrefix string) ([]string, error) {
	s.access.Lock()
	defer s.access.Unlock()
	var ret []string
	for key := range s.values {
		if strings.HasPrefix(key, keyPrefix) {
			ret = append(ret, key)
		}
	}
	return ret, nil
}

func (s *scopedTransientStorageImpl) Clear(ctx context.Context) {
	s.access.Lock()
	defer s.access.Unlock()
	for _, v := range s.values {
		if sw, ok := v.(storage.TransientStorageLifecycleReceiver); ok {
			_ = sw.Close()
		}
	}
	s.values = map[string]interface{}{}
	for _, v := range s.scopes {
		v.Clear(ctx)
	}
	s.scopes = map[string]storage.ScopedTransientStorage{}
}

func (s *scopedTransientStorageImpl) NarrowScope(ctx context.Context, key string) (storage.ScopedTransientStorage, error) {
	s.access.Lock()
	defer s.access.Unlock()
	sw, ok := s.scopes[key]
	if !ok {
		scope := NewScopedTransientStorageImpl()
		s.scopes[key] = scope
		return scope, nil
	}
	return sw, nil
}

func (s *scopedTransientStorageImpl) DropScope(ctx context.Context, key string) error {
	s.access.Lock()
	defer s.access.Unlock()
	if v, ok := s.scopes[key]; ok {
		v.Clear(ctx)
	}
	delete(s.scopes, key)
	return nil
}
