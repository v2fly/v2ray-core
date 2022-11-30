package mergers

import (
	"io"
	"path/filepath"
	"strings"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
)

// MergeAs load input and merge as specified format into m
func MergeAs(formatName string, input interface{}, m map[string]interface{}) error {
	f, found := mergersByName[formatName]
	if !found {
		return newError("format merger not found for: ", formatName)
	}
	return f.Merge(input, m)
}

// Merge loads inputs and merges them into m
// it detects extension for merger selecting, or try all mergers
// if no extension found
func Merge(input interface{}, m map[string]interface{}) error {
	if input == nil {
		return nil
	}
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
		// read to []byte incase it tries different mergers
		bs, err := io.ReadAll(v)
		if err != nil {
			return err
		}
		err = mergeSingleFile(bs, m)
		if err != nil {
			return err
		}
	default:
		return newError("unknown merge input type")
	}
	return nil
}

func mergeSingleFile(input interface{}, m map[string]interface{}) error {
	if file, ok := input.(string); ok {
		ext := getExtension(file)
		if ext != "" {
			lext := strings.ToLower(ext)
			f, found := mergersByExt[lext]
			if !found {
				return newError("unmergeable format extension: ", ext)
			}
			return f.Merge(file, m)
		}
	}
	// no extension, try all mergers
	for _, f := range mergersByName {
		if f.Name == core.FormatAuto {
			continue
		}
		err := f.Merge(input, m)
		if err == nil {
			return nil
		}
	}
	return newError("tried all mergers but failed for: ", input).AtWarning()
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}
