package server

func newEmptyCipherSuiteLookuper() *ciphersuiteLookuper {
	return &ciphersuiteLookuper{
		ciphersuiteMap: make(map[uint16]bool),
	}
}

func newCipherSuiteLookuperFromUint32Array(source []uint32) (*ciphersuiteLookuper, error) {
	if len(source) == 0 {
		return nil, newError("ciphersuite list is empty")
	}
	ciphersuiteUint16Array := make([]uint16, len(source))
	for i, ciphersuite := range source {
		if ciphersuite > 0xFFFF {
			return nil, newError("ciphersuite value out of range: ", ciphersuite)
		}
		ciphersuiteUint16Array[i] = uint16(ciphersuite)
	}
	return newCipherSuiteLookuperFromUint16Array(ciphersuiteUint16Array), nil
}

func newCipherSuiteLookuperFromUint16Array(source []uint16) *ciphersuiteLookuper {
	ciphersuiteMap := make(map[uint16]bool, len(source))
	for _, ciphersuite := range source {
		ciphersuiteMap[ciphersuite] = true
	}
	return &ciphersuiteLookuper{
		ciphersuiteMap: ciphersuiteMap,
	}
}

type ciphersuiteLookuper struct {
	ciphersuiteMap map[uint16]bool
}

func (l *ciphersuiteLookuper) Lookup(ciphersuite uint16) bool {
	if result, ok := l.ciphersuiteMap[ciphersuite]; ok {
		return result
	}
	return false
}
