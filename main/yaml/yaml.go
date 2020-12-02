package yaml

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf/json"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{"YAML"},
		Extension: []string{".yml", ".yaml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				bs, err := yamlsToJSONs(v)
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

func yamlsToJSONs(files []string) ([][]byte, error) {
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
