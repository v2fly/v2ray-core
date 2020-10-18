package yaml

import (
	"io"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/main/confloader"
)

//go:generate go run v2ray.com/core/common/errors/errorgen

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "YAML",
		Extension: []string{"yml", "yaml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				r, err := confloader.LoadExtConfig(v)
				if err != nil {
					return nil, newError("failed to execute v2ctl to convert config file.").Base(err).AtWarning()
				}
				return core.LoadConfig("protobuf", "", r)
			case io.Reader:
				return serial.LoadYAMLConfig(v)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
