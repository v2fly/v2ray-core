// +build windows,android

package platform

import	"path/filepath"

// GetAssetLocation search for `file` in the excutable dir
func GetAssetLocation(file string) string {
	const name = "v2ray.location.asset"
	assetPath := NewEnvFlag(name).GetValue(getExecutableDir)
	return filepath.Join(assetPath, file)
}
