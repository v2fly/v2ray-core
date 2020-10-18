package serial

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"v2ray.com/core"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/infra/conf"
	json_reader "v2ray.com/core/infra/conf/json"
)

type offset struct {
	line int
	char int
}

func findOffset(b []byte, o int) *offset {
	if o >= len(b) || o < 0 {
		return nil
	}

	line := 1
	char := 0
	for i, x := range b {
		if i == o {
			break
		}
		if x == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}

	return &offset{line: line, char: char}
}

// DecodeJSONConfig reads from reader and decode the config into *conf.Config
// syntax error could be detected.
func DecodeJSONConfig(reader io.Reader) (*conf.Config, error) {
	jsonConfig := &conf.Config{}

	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&json_reader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(jsonConfig); err != nil {
		var pos *offset
		cause := errors.Cause(err)
		switch tErr := cause.(type) {
		case *json.SyntaxError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		case *json.UnmarshalTypeError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		}
		if pos != nil {
			return nil, newError("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return nil, newError("failed to read config file").Base(err)
	}

	return jsonConfig, nil
}

// DecodeYAMLConfig reads from reader and decode the config into *conf.Config
// syntax error could NOT be detected.
func DecodeYAMLConfig(reader io.Reader) (*conf.Config, error) {
	yamlConfig := &conf.Config{}

	// Since v2ray use json.RawMessage a lot, so we use a wrapper to convert yaml to json
	if tmpBuf, err := ioutil.ReadAll(reader); err != nil {
		return nil, newError("failed to read").Base(err)
	} else if err := yaml.Unmarshal(tmpBuf, yamlConfig); err != nil {
		return nil, newError("failed to parse config").Base(err)
	}

	return yamlConfig, nil
}

// LoadJSONConfig uses content in reader to return a core.Config
func LoadJSONConfig(reader io.Reader) (*core.Config, error) {
	jsonConfig, err := DecodeJSONConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := jsonConfig.Build()
	if err != nil {
		return nil, newError("failed to parse json config").Base(err)
	}

	return pbConfig, nil
}

// LoadYAMLConfig uses content in reader to return a core.Config
func LoadYAMLConfig(reader io.Reader) (*core.Config, error) {
	yamlConfig, err := DecodeYAMLConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := yamlConfig.Build()
	if err != nil {
		return nil, newError("failed to parse yaml config").Base(err)
	}

	return pbConfig, nil
}
