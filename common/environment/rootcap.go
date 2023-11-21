package environment

type RootEnvironment interface {
	AppEnvironment(tag string) AppEnvironment
	ProxyEnvironment(tag string) ProxyEnvironment
	DropProxyEnvironment(tag string) error
	doNotImpl()
}
