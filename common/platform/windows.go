// +build windows

package platform

import "path/filepath"

func ExpandEnv(s string) string {
	// TODO
	return s
}

func LineSeparator() string {
	return "\r\n"
}

func GetToolLocation(file string) string {
	const name = "v2ray.location.tool"
	toolPath := EnvFlag{Name: name, AltName: NormalizeEnvName(name)}.GetValue(getExecutableDir)
	return filepath.Join(toolPath, file+".exe")
}

// GetAssetLocation search for `file` in the executable dir
func GetAssetLocation(file string) string {
	filepathCleaned := filepath.Clean(file)
	if strings.HasPrefix("..", filepathCleaned) {
		newError("directory transversal is not allowed for assets. This will be forbidden in v5.").AtWarning().WriteToLog()
	}
	const name = "v2ray.location.asset"
	assetPath := NewEnvFlag(name).GetValue(getExecutableDir)
	return filepath.Join(assetPath, file)
}
