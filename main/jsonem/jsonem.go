package jsonem

import (
	"bytes"
	"io"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf/merge"
	"v2ray.com/core/infra/conf/serial"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{"JSON"},
		Extension: []string{".json", ".jsonc"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				data, err := merge.ToJSON(v)
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
				return serial.LoadJSONConfig(v)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
