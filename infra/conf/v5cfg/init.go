package v5cfg

import (
	"bytes"
	"io"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/cmdarg"
	"github.com/ghxhy/v2ray-core/v5/infra/conf/json"
)

const jsonV5 = "jsonv5"

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{jsonV5},
		Extension: []string{".v5.json", ".v5.jsonc"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case string:
				r, err := cmdarg.LoadArg(v)
				if err != nil {
					return nil, err
				}
				data, err := buf.ReadAllToBytes(&json.Reader{
					Reader: r,
				})
				if err != nil {
					return nil, err
				}
				return loadJSONConfig(data)
			case []byte:
				r := &json.Reader{
					Reader: bytes.NewReader(v),
				}
				data, err := buf.ReadAllToBytes(r)
				if err != nil {
					return nil, err
				}
				return loadJSONConfig(data)
			case io.Reader:
				data, err := buf.ReadAllToBytes(&json.Reader{
					Reader: v,
				})
				if err != nil {
					return nil, err
				}
				return loadJSONConfig(data)
			default:
				return nil, newError("unknown type")
			}
		},
	}))
}
