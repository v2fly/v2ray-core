package yaml

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf/json"
	"v2ray.com/core/infra/conf/merge"
	"v2ray.com/core/infra/conf/serial"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{"YAML"},
		Extension: []string{".yml", ".yaml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				bs, err := yamlsToJSON(v)
				if err != nil {
					return nil, err
				}
				data, err := merge.BytesToJSON(bs)
				if err != nil {
					return nil, err
				}
				r := bytes.NewReader(data)
				cf, err := serial.DecodeJSONConfig(r)
				if err != nil {
					return nil, err
				}
				return cf.Build()
			case io.Reader:
				bs, err := ioutil.ReadAll(v)
				if err != nil {
					return nil, err
				}
				bs, err = json.FromYAML(bs)
				if err != nil {
					return nil, err
				}
				return serial.LoadJSONConfig(bytes.NewBuffer(bs))
			default:
				return nil, errors.New("unknow type")
			}
		},
	}))
}

func yamlsToJSON(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := json.FromYAML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}
