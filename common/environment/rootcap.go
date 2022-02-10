package environment

type RootEnvironment interface {
	AppEnvironment(tag string) AppEnvironment
	ProxyEnvironment(tag string) ProxyEnvironment
	doNotImpl()
}
