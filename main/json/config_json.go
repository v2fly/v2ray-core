package json

//go:generate go run github.com/v2fly/v2ray-core/common/errors/errorgen

import (
	"io"

	core "github.com/v2fly/v2ray-core"
	"github.com/v2fly/v2ray-core/common"
	"github.com/v2fly/v2ray-core/common/cmdarg"
	"github.com/v2fly/v2ray-core/infra/conf/serial"
	"github.com/v2fly/v2ray-core/main/confloader"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "JSON",
		Extension: []string{"json"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				r, err := confloader.LoadExtConfig(v)
				if err != nil {
					return nil, newError("failed to execute v2ctl to convert config file.").Base(err).AtWarning()
				}
				return core.LoadConfig("protobuf", "", r)
			case io.Reader:
				return serial.LoadJSONConfig(v)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
