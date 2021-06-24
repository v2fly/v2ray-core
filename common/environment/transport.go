package environment

import "github.com/v2fly/v2ray-core/v4/features/extension"

type TransportEnvironmentCapacitySet interface {
	BaseEnvironmentCapabilitySet
	SystemNetworkCapabilitySet
	InstanceNetworkCapabilitySet

	TransientStorage() extension.ScopedTransientStorage
}

type TransportEnvironment interface {
	TransportEnvironmentCapacitySet

	NarrowScope(key []byte) (TransportEnvironment, error)
	doNotImpl()
}
