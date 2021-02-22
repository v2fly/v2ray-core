package mergers

import "strings"

// GetExtensions get extensions of given format
func GetExtensions(formatName string) ([]string, error) {
	lowerName := strings.ToLower(formatName)
	if lowerName == "auto" {
		return GetAllExtensions(), nil
	}
	f, found := mergersByName[lowerName]
	if !found {
		return nil, newError(formatName+" not found", formatName).AtWarning()
	}
	return f.Extensions, nil
}

// GetAllExtensions get all extensions supported
func GetAllExtensions() []string {
	extensions := make([]string, 0)
	for _, f := range mergersByName {
		extensions = append(extensions, f.Extensions...)
	}
	return extensions
}
