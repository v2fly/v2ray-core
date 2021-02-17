package core

import (
	"io"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
)

// ConfigFormat is a configurable format of V2Ray config file.
type ConfigFormat struct {
	Name      []string
	Extension []string
	Loader    ConfigLoader
}

// ConfigLoader is a utility to load V2Ray config from external source.
type ConfigLoader func(input interface{}) (*Config, error)

var (
	configLoaderByName = make(map[string]*ConfigFormat)
	configLoaderByExt  = make(map[string]*ConfigFormat)
)

// RegisterConfigLoader add a new ConfigLoader.
func RegisterConfigLoader(format *ConfigFormat) error {
	for _, name := range format.Name {
		lname := strings.ToLower(name)
		if _, found := configLoaderByName[lname]; found {
			return newError(name, " already registered.")
		}
		configLoaderByName[lname] = format
	}

	for _, ext := range format.Extension {
		lext := strings.ToLower(ext)
		if f, found := configLoaderByExt[lext]; found {
			return newError(ext, " already registered to ", f.Name)
		}
		configLoaderByExt[lext] = format
	}

	return nil
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// GetConfigLoader get config loader by name and filename.
// Specify formatName to explicitly select a loader.
// Specify filename to choose loader by detect its extension.
// Leave formatName and filename blank for default loader
func GetConfigLoader(formatName string, filename string) (*ConfigFormat, error) {
	if formatName != "" {
		// if explicitly specified, we can safely assume that user knows what they are
		if f, found := configLoaderByName[formatName]; found {
			return f, nil
		}
		return nil, newError("Unable to load config in ", formatName).AtWarning()
	}
	// no explicitly specified loader, extenstion detect first
	if ext := getExtension(filename); len(ext) > 0 {
		if f, found := configLoaderByExt[ext]; found {
			return f, nil
		}
	}
	// default loader
	if f, found := configLoaderByName["json"]; found {
		return f, nil
	}
	panic("default loader not found")
}

// LoadConfig loads config with given format from given source.
// input accepts 2 different types:
// * []string slice of multiple filename/url(s) to open to read
// * io.Reader that reads a config content (the original way)
func LoadConfig(formatName string, filename string, input interface{}) (*Config, error) {
	f, err := GetConfigLoader(formatName, filename)
	if err != nil {
		return nil, err
	}
	return f.Loader(input)
}

func loadProtobufConfig(data []byte) (*Config, error) {
	config := new(Config)
	if err := proto.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func init() {
	common.Must(RegisterConfigLoader(&ConfigFormat{
		Name:      []string{"Protobuf", "pb"},
		Extension: []string{".pb"},
		Loader: func(input interface{}) (*Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				r, err := cmdarg.LoadArg(v[0])
				if err != nil {
					return nil, err
				}
				data, err := buf.ReadAllToBytes(r)
				if err != nil {
					return nil, err
				}
				return loadProtobufConfig(data)
			case io.Reader:
				data, err := buf.ReadAllToBytes(v)
				if err != nil {
					return nil, err
				}
				return loadProtobufConfig(data)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
