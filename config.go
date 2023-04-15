package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
)

const (
	// FormatAuto represents all available formats by auto selecting
	FormatAuto = "auto"
	// FormatJSON represents json format
	FormatJSON = "json"
	// FormatTOML represents toml format
	FormatTOML = "toml"
	// FormatYAML represents yaml format
	FormatYAML = "yaml"
	// FormatProtobuf represents protobuf format
	FormatProtobuf = "protobuf"
	// FormatProtobufShort is the short of FormatProtobuf
	FormatProtobufShort = "pb"
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
	configLoaders      = make([]*ConfigFormat, 0)
	configLoaderByName = make(map[string]*ConfigFormat)
	configLoaderByExt  = make(map[string]*ConfigFormat)
)

// RegisterConfigLoader add a new ConfigLoader.
func RegisterConfigLoader(format *ConfigFormat) error {
	for _, name := range format.Name {
		if _, found := configLoaderByName[name]; found {
			return newError(name, " already registered.")
		}
		configLoaderByName[name] = format
	}

	for _, ext := range format.Extension {
		lext := strings.ToLower(ext)
		if f, found := configLoaderByExt[lext]; found {
			return newError(ext, " already registered to ", f.Name)
		}
		configLoaderByExt[lext] = format
	}
	configLoaders = append(configLoaders, format)
	return nil
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// GetLoaderExtensions get config loader extensions.
func GetLoaderExtensions(formatName string) ([]string, error) {
	if formatName == FormatAuto {
		return GetAllExtensions(), nil
	}
	if f, found := configLoaderByName[formatName]; found {
		return f.Extension, nil
	}
	return nil, newError("config loader not found: ", formatName).AtWarning()
}

// GetAllExtensions get all extensions supported
func GetAllExtensions() []string {
	extensions := make([]string, 0)
	for _, f := range configLoaderByName {
		extensions = append(extensions, f.Extension...)
	}
	return extensions
}

// LoadConfig loads multiple config with given format from given source.
// input accepts:
// * string of a single filename/url(s) to open to read
// * []string slice of multiple filename/url(s) to open to read
// * io.Reader that reads a config content (the original way)
func LoadConfig(formatName string, input interface{}) (*Config, error) {
	cnt := getInputCount(input)
	if cnt == 0 {
		log.Println("Using config from STDIN")
		input = os.Stdin
		cnt = 1
	}
	if formatName == FormatAuto && cnt == 1 {
		// This ensures only to call auto loader for multiple files,
		// so that it can only care about merging scenarios
		return loadSingleConfigAutoFormat(input)
	}
	// if input is a slice with single element, extract it
	// so that unmergeable loaders don't need to deal with
	// slices
	s := reflect.Indirect(reflect.ValueOf(input))
	k := s.Kind()
	if (k == reflect.Slice || k == reflect.Array) && s.Len() == 1 {
		value := reflect.Indirect(s.Index(0))
		if value.Kind() == reflect.String {
			// string type alias
			input = fmt.Sprint(value.Interface())
		} else {
			input = value.Interface()
		}
	}
	f, found := configLoaderByName[formatName]
	if !found {
		return nil, newError("config loader not found: ", formatName).AtWarning()
	}
	return f.Loader(input)
}

// loadSingleConfigAutoFormat loads a single config with from given source.
// input accepts:
// * string of a single filename/url(s) to open to read
// * io.Reader that reads a config content (the original way)
func loadSingleConfigAutoFormat(input interface{}) (*Config, error) {
	switch v := input.(type) {
	case cmdarg.Arg:
		return loadSingleConfigAutoFormatFromFile(v.String())
	case string:
		return loadSingleConfigByTryingAllLoaders(v)
	case io.Reader:
		data, err := buf.ReadAllToBytes(v)
		if err != nil {
			return nil, err
		}
		return loadSingleConfigByTryingAllLoaders(data)
	default:
		return loadSingleConfigByTryingAllLoaders(v)
	}
}

func loadSingleConfigAutoFormatFromFile(file string) (*Config, error) {
	extension := getExtension(file)
	if extension != "" {
		lowerName := strings.ToLower(extension)
		if f, found := configLoaderByExt[lowerName]; found {
			return f.Loader(file)
		}
		return nil, newError("config loader not found for: ", extension).AtWarning()
	}

	return loadSingleConfigByTryingAllLoaders(file)
}

func loadSingleConfigByTryingAllLoaders(input interface{}) (*Config, error) {
	var errorReasons strings.Builder

	for _, f := range configLoaders {
		if f.Name[0] == FormatAuto {
			continue
		}
		c, err := f.Loader(input)
		if err == nil {
			return c, nil
		}
		errorReasons.WriteString(fmt.Sprintf("unable to parse as %v:%v;", f.Name[0], err.Error()))
	}

	return nil, newError("tried all loaders but failed when attempting to parse: ", input, ";", errorReasons.String()).AtWarning()
}

func getInputCount(input interface{}) int {
	s := reflect.Indirect(reflect.ValueOf(input))
	k := s.Kind()
	if k == reflect.Slice || k == reflect.Array {
		return s.Len()
	}
	return 1
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
		Name:      []string{FormatProtobuf, FormatProtobufShort},
		Extension: []string{".pb"},
		Loader: func(input interface{}) (*Config, error) {
			switch v := input.(type) {
			case string:
				r, err := cmdarg.LoadArg(v)
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
				return nil, newError("unknown type")
			}
		},
	}))
}
