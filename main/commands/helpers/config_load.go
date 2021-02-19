package helpers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"github.com/v2fly/v2ray-core/v4/infra/conf"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/infra/conf/mergers"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
)

// LoadConfig load config files to *conf.Config, it will:
// - resolve folder to files
// - try to read stdin if no file specified
func LoadConfig(files []string, format string, recursively bool) (*conf.Config, error) {
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
	var (
		stdin []byte
		err   error
	)
	if len(files) > 0 {
		var extensions []string
		extensions, err = mergers.GetExtensions(format)
		if err != nil {
			return nil, err
		}
		files, err = ResolveFolderToFiles(files, extensions, recursively)
		if err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		stdin, err = readStdin()
		if err != nil {
			return nil, err
		}
		if len(stdin) == 0 {
			return nil, errors.New("no config found")
		}
	}
	m := make(map[string]interface{})
	if stdin != nil {
		err = mergers.MergeAs(format, stdin, m)
	} else {
		err = mergers.MergeAs(format, files, m)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func readStdin() ([]byte, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		return nil, nil
	}
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return stdin, nil
}
