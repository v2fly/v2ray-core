package securedload

func GetAssetSecured(name string) ([]byte, error) {
	var err error
	for k, v := range knownProtectedLoader {
		if loadedData, errLoad := v.VerifyAndLoad(name); errLoad == nil {
			return loadedData, nil
		} else {
			err = newError(k, " is not loading executable file").Base(errLoad)
		}
	}
	return nil, err
}
