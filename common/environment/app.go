package environment

type AppEnvironmentCapabilitySet interface {
	BaseEnvironmentCapabilitySet
	SystemNetworkCapabilitySet
	InstanceNetworkCapabilitySet
	FileSystemCapabilitySet
	PersistentStorageCapabilitySet
	TransientStorageCapabilitySet
}

type AppEnvironment interface {
	AppEnvironmentCapabilitySet
	NarrowScope(key string) (AppEnvironment, error)
	doNotImpl()
}
