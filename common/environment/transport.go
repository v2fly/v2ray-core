package environment

type TransportEnvironmentCapacitySet interface {
	BaseEnvironmentCapabilitySet
	SystemNetworkCapabilitySet
	InstanceNetworkCapabilitySet
	TransientStorageCapabilitySet
	ProxyMetadataCapabilitySet
}

type TransportEnvironment interface {
	TransportEnvironmentCapacitySet
	NarrowScope(key string) (TransportEnvironment, error)
	doNotImpl()
}
