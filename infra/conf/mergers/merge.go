package mergers

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
)

// MergeAs load input and merge as specified format into m
func MergeAs(formatName string, input interface{}, m map[string]interface{}) error {
	f, found := mergeLoaderByName[formatName]
	if !found {
		return newError("format loader not found for: ", formatName)
	}
	return f.Loader(input, m)
}

// Merge loads inputs and merges them into m
// it detects extension for loader selecting, or try all loaders
// if no extension found
func Merge(input interface{}, m map[string]interface{}) error {
	switch v := input.(type) {
	case string:
		err := mergeSingleFile(v, m)
		if err != nil {
			return err
		}
	case []string:
		for _, file := range v {
			err := mergeSingleFile(file, m)
			if err != nil {
				return err
			}
		}
	case cmdarg.Arg:
		for _, file := range v {
			err := mergeSingleFile(file, m)
			if err != nil {
				return err
			}
		}
	case []byte:
		err := mergeSingleFile(v, m)
		if err != nil {
			return err
		}
	case io.Reader:
		// read to []byte incase it tries different loaders
		bs, err := ioutil.ReadAll(v)
		if err != nil {
			return err
		}
		err = mergeSingleFile(bs, m)
		if err != nil {
			return err
		}
	default:
		return newError("unknow merge input type")
	}
	return nil
}

func mergeSingleFile(input interface{}, m map[string]interface{}) error {
	if file, ok := input.(string); ok {
		ext := getExtension(file)
		if ext != "" {
			lext := strings.ToLower(ext)
			f, found := mergeLoaderByExt[lext]
			if !found {
				return newError("unmergeable format extension: ", ext)
			}
			return f.Loader(file, m)
		}
	}
	// no extension, try all loaders
	for _, f := range mergeLoaderByName {
		if f.Name == core.FormatAuto {
			continue
		}
		err := f.Loader(input, m)
		if err == nil {
			return nil
		}
	}
	return newError("tried all loaders but failed for: ", input).AtWarning()
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}
