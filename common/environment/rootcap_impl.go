package environment

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/platform/filesystem/fsifce"
	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tagged"
)

func NewRootEnvImpl(ctx context.Context, transientStorage storage.ScopedTransientStorage,
	systemDialer internet.SystemDialer, systemListener internet.SystemListener,
	filesystem FileSystemCapabilitySet, persistStorage storage.ScopedPersistentStorage,
) RootEnvironment {
	return &rootEnvImpl{
		transientStorage: transientStorage,
		systemListener:   systemListener,
		systemDialer:     systemDialer,
		filesystem:       filesystem,
		persistStorage:   persistStorage,
		ctx:              ctx,
	}
}

type rootEnvImpl struct {
	persistStorage   storage.ScopedPersistentStorage
	transientStorage storage.ScopedTransientStorage
	systemDialer     internet.SystemDialer
	systemListener   internet.SystemListener
	filesystem       FileSystemCapabilitySet

	ctx context.Context
}

func (r *rootEnvImpl) doNotImpl() {
	panic("placeholder doNotImpl")
}

func (r *rootEnvImpl) AppEnvironment(tag string) AppEnvironment {
	transientStorage, err := r.transientStorage.NarrowScope(r.ctx, tag)
	if err != nil {
		return nil
	}
	persistStorage, err := r.persistStorage.NarrowScope(r.ctx, []byte(tag))
	if err != nil {
		return nil
	}
	return &appEnvImpl{
		transientStorage: transientStorage,
		persistStorage:   persistStorage,
		systemListener:   r.systemListener,
		systemDialer:     r.systemDialer,
		filesystem:       r.filesystem,
		ctx:              r.ctx,
	}
}

func (r *rootEnvImpl) ProxyEnvironment(tag string) ProxyEnvironment {
	transientStorage, err := r.transientStorage.NarrowScope(r.ctx, tag)
	if err != nil {
		return nil
	}
	return &proxyEnvImpl{
		transientStorage: transientStorage,
		systemListener:   r.systemListener,
		systemDialer:     r.systemDialer,
		ctx:              r.ctx,
	}
}

func (r *rootEnvImpl) DropProxyEnvironment(tag string) error {
	transientStorage, err := r.transientStorage.NarrowScope(r.ctx, tag)
	if err != nil {
		return err
	}
	transientStorage.Clear(r.ctx)
	return r.transientStorage.DropScope(r.ctx, tag)
}

type appEnvImpl struct {
	persistStorage   storage.ScopedPersistentStorage
	transientStorage storage.ScopedTransientStorage
	systemDialer     internet.SystemDialer
	systemListener   internet.SystemListener
	filesystem       FileSystemCapabilitySet

	ctx context.Context
}

func (a *appEnvImpl) RequireFeatures() interface{} {
	panic("implement me")
}

func (a *appEnvImpl) RecordLog() interface{} {
	panic("implement me")
}

func (a *appEnvImpl) Dialer() internet.SystemDialer {
	panic("implement me")
}

func (a *appEnvImpl) Listener() internet.SystemListener {
	panic("implement me")
}

func (a *appEnvImpl) OutboundDialer() tagged.DialFunc {
	return internet.DialTaggedOutbound
}

func (a *appEnvImpl) OpenFileForReadSeek() fsifce.FileSeekerFunc {
	return a.filesystem.OpenFileForReadSeek()
}

func (a *appEnvImpl) OpenFileForRead() fsifce.FileReaderFunc {
	return a.filesystem.OpenFileForRead()
}

func (a *appEnvImpl) OpenFileForWrite() fsifce.FileWriterFunc {
	return a.filesystem.OpenFileForWrite()
}

func (a *appEnvImpl) ReadDir() fsifce.FileReadDirFunc {
	return a.filesystem.ReadDir()
}

func (a *appEnvImpl) RemoveFile() fsifce.FileRemoveFunc {
	return a.filesystem.RemoveFile()
}

func (a *appEnvImpl) PersistentStorage() storage.ScopedPersistentStorage {
	return a.persistStorage
}

func (a *appEnvImpl) TransientStorage() storage.ScopedTransientStorage {
	return a.transientStorage
}

func (a *appEnvImpl) NarrowScope(key string) (AppEnvironment, error) {
	transientStorage, err := a.transientStorage.NarrowScope(a.ctx, key)
	if err != nil {
		return nil, err
	}
	return &appEnvImpl{
		transientStorage: transientStorage,
		systemDialer:     a.systemDialer,
		systemListener:   a.systemListener,
		ctx:              a.ctx,
	}, nil
}

func (a *appEnvImpl) doNotImpl() {
	panic("placeholder doNotImpl")
}

type proxyEnvImpl struct {
	transientStorage storage.ScopedTransientStorage
	systemDialer     internet.SystemDialer
	systemListener   internet.SystemListener

	scopeName string

	ctx context.Context
}

func (p *proxyEnvImpl) RequireFeatures() interface{} {
	panic("implement me")
}

func (p *proxyEnvImpl) RecordLog() interface{} {
	panic("implement me")
}

func (p *proxyEnvImpl) OutboundDialer() tagged.DialFunc {
	panic("implement me")
}

func (p *proxyEnvImpl) TransientStorage() storage.ScopedTransientStorage {
	return p.transientStorage
}

func (p *proxyEnvImpl) SelfProxyTag() string {
	return p.scopeName
}

func (p *proxyEnvImpl) NarrowScope(key string) (ProxyEnvironment, error) {
	transientStorage, err := p.transientStorage.NarrowScope(p.ctx, key)
	if err != nil {
		return nil, err
	}
	return &proxyEnvImpl{
		transientStorage: transientStorage,
		scopeName:        p.scopeName,
		ctx:              p.ctx,
	}, nil
}

func (p *proxyEnvImpl) NarrowScopeToTransport(key string) (TransportEnvironment, error) {
	transientStorage, err := p.transientStorage.NarrowScope(p.ctx, key)
	if err != nil {
		return nil, err
	}
	return &transportEnvImpl{
		ctx:              p.ctx,
		transientStorage: transientStorage,
		systemDialer:     p.systemDialer,
		systemListener:   p.systemListener,
		selfProxyTag:     p.scopeName,
	}, nil
}

func (p *proxyEnvImpl) doNotImpl() {
	panic("placeholder doNotImpl")
}

type transportEnvImpl struct {
	transientStorage storage.ScopedTransientStorage
	systemDialer     internet.SystemDialer
	systemListener   internet.SystemListener

	ctx context.Context

	selfProxyTag string
}

func (t *transportEnvImpl) RequireFeatures() interface{} {
	panic("implement me")
}

func (t *transportEnvImpl) SelfProxyTag() string {
	return t.selfProxyTag
}

func (t *transportEnvImpl) RecordLog() interface{} {
	panic("implement me")
}

func (t *transportEnvImpl) Dialer() internet.SystemDialer {
	return t.systemDialer
}

func (t *transportEnvImpl) Listener() internet.SystemListener {
	return t.systemListener
}

func (t *transportEnvImpl) OutboundDialer() tagged.DialFunc {
	return tagged.Dialer
}

func (t *transportEnvImpl) TransientStorage() storage.ScopedTransientStorage {
	return t.transientStorage
}

func (t *transportEnvImpl) NarrowScope(key string) (TransportEnvironment, error) {
	transientStorage, err := t.transientStorage.NarrowScope(t.ctx, key)
	if err != nil {
		return nil, err
	}
	return &transportEnvImpl{
		ctx:              t.ctx,
		transientStorage: transientStorage,
	}, nil
}

func (t *transportEnvImpl) doNotImpl() {
	panic("implement me")
}
