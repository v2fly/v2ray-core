package environment

import (
	"github.com/v2fly/v2ray-core/v4/features/extension/storage"
)

type AppEnvironmentCapabilitySet interface {
	BaseEnvironmentCapabilitySet
	SystemNetworkCapabilitySet
	InstanceNetworkCapabilitySet
	FileSystemCapabilitySet

	PersistentStorage() storage.ScopedPersistentStorage
	TransientStorage() storage.ScopedTransientStorage
}

type AppEnvironment interface {
	AppEnvironmentCapabilitySet

	NarrowScope(key []byte) (AppEnvironment, error)
	doNotImpl()
}
