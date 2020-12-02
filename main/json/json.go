package json

import (
	"bytes"
	"io"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{"JSON"},
		Extension: []string{".json", ".jsonc"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				data, err := merge.FilesToJSON(v)
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
