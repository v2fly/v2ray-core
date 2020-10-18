package yamlem

import (
	"io"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/main/confloader"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "YAML",
		Extension: []string{"yml", "yaml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				cf := &conf.Config{}
				for _, arg := range v {
					newError("Reading config: ", arg).AtInfo().WriteToLog()
					r, err := confloader.LoadConfig(arg)
					common.Must(err)
					c, err := serial.DecodeYAMLConfig(r)
					common.Must(err)
					cf.Override(c, arg)
				}
				return cf.Build()
			case io.Reader:
				return serial.LoadYAMLConfig(v)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
