package helpers

import (
	"bytes"
	"os"

	"github.com/v2fly/v2ray-core/v5/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v5/infra/conf/mergers"
	"github.com/v2fly/v2ray-core/v5/infra/conf/serial"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
)

// LoadConfig load config files to *conf.Config, it will:
// - resolve folder to files
// - try to read stdin if no file specified
func LoadConfig(files []string, format string, recursively bool) (*v4.Config, error) {
	m, err := LoadConfigToMap(files, format, recursively)
	if err != nil {
		return nil, err
	}
	bs, err := merge.FromMap(m)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(bs)
	return serial.DecodeJSONConfig(r)
}

// LoadConfigToMap load config files to map, it will:
// - resolve folder to files
// - try to read stdin if no file specified
func LoadConfigToMap(files []string, format string, recursively bool) (map[string]interface{}, error) {
	var err error
	if len(files) > 0 {
		var extensions []string
		extensions, err := mergers.GetExtensions(format)
		if err != nil {
			return nil, err
		}
		files, err = ResolveFolderToFiles(files, extensions, recursively)
		if err != nil {
			return nil, err
		}
	}
	m := make(map[string]interface{})
	if len(files) == 0 {
		err = mergers.MergeAs(format, os.Stdin, m)
	} else {
		err = mergers.MergeAs(format, files, m)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}
