package mergers

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"strings"
)

// MergeableFormat is a configurable mergeable format of V2Ray config file.
type MergeableFormat struct {
	Name       string
	Extensions []string
	Loader     MergeLoader
}

// MergeLoader is a utility to merge V2Ray config from external source into a map and returns it.
type MergeLoader func(input interface{}, m map[string]interface{}) error

var (
	mergeLoaderByName = make(map[string]*MergeableFormat)
	mergeLoaderByExt  = make(map[string]*MergeableFormat)
)

// RegisterMergeLoader add a new MergeLoader.
func RegisterMergeLoader(format *MergeableFormat) error {
	if _, found := mergeLoaderByName[format.Name]; found {
		return newError(format.Name, " already registered.")
	}
	mergeLoaderByName[format.Name] = format

	for _, ext := range format.Extensions {
		lext := strings.ToLower(ext)
		if f, found := mergeLoaderByExt[lext]; found {
			return newError(ext, " already registered to ", f.Name)
		}
		mergeLoaderByExt[lext] = format
	}

	return nil
}
