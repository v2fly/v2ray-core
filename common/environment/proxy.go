package environment

type ProxyEnvironmentCapabilitySet interface {
	BaseEnvironmentCapabilitySet
	InstanceNetworkCapabilitySet
	TransientStorageCapabilitySet
	ProxyMetadataCapabilitySet
}

// TODO Add NarrowScopeToConnection

type ProxyEnvironment interface {
	ProxyEnvironmentCapabilitySet
	NarrowScope(key string) (ProxyEnvironment, error)
	NarrowScopeToTransport(key string) (TransportEnvironment, error)
	doNotImpl()
}
