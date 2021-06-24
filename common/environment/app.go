package environment

import (
	"github.com/v2fly/v2ray-core/v4/features/extension"
)

type AppEnvironmentCapabilitySet interface {
	BaseEnvironmentCapabilitySet
	SystemNetworkCapabilitySet
	InstanceNetworkCapabilitySet
	FileSystemCapabilitySet

	PersistentStorage() extension.ScopedPersistentStorage
	TransientStorage() extension.ScopedTransientStorage
}

type AppEnvironment interface {
	AppEnvironmentCapabilitySet

	NarrowScope(key []byte) (AppEnvironment, error)
	doNotImpl()
}
