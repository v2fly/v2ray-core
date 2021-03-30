package securedload

type ProtectedLoader interface {
	VerifyAndLoad(filename string) ([]byte, error)
}

var knownProtectedLoader map[string]ProtectedLoader

func RegisterProtectedLoader(name string, sv ProtectedLoader) {
	if knownProtectedLoader == nil {
		knownProtectedLoader = map[string]ProtectedLoader{}
	}
	knownProtectedLoader[name] = sv
}
