package environment

import "github.com/v2fly/v2ray-core/v4/features/extension"

type ProxyEnvironmentCapabilitySet interface {
	BaseEnvironmentCapabilitySet
	InstanceNetworkCapabilitySet

	TransientStorage() extension.ScopedTransientStorage
}

type ProxyEnvironment interface {
	ProxyEnvironmentCapabilitySet

	NarrowScope(key []byte) (ProxyEnvironment, error)
	NarrowScopeToTransport(key []byte) (TransportEnvironment, error)

	doNotImpl()
}
