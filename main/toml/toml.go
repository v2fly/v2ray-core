package toml

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
		Name:      []string{"TOML"},
		Extension: []string{".toml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				bs, err := tomlsToJSONs(v)
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
				bs, err = json.FromTOML(bs)
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

func tomlsToJSONs(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := json.FromTOML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}
