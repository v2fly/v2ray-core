//go:build windows
// +build windows

package platform

import "path/filepath"

func LineSeparator() string {
	return "\r\n"
}

// GetAssetLocation search for `file` in the excutable dir
func GetAssetLocation(file string) string {
	const name = "v2ray.location.asset"
	assetPath := NewEnvFlag(name).GetValue(getExecutableDir)
	return filepath.Join(assetPath, file)
}
