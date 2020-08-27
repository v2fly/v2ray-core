package jsonem

import (
	"io"

	core "github.com/v2fly/v2ray-core"
	"github.com/v2fly/v2ray-core/common"
	"github.com/v2fly/v2ray-core/common/cmdarg"
	"github.com/v2fly/v2ray-core/infra/conf"
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
				cf := &conf.Config{}
				for _, arg := range v {
					newError("Reading config: ", arg).AtInfo().WriteToLog()
					r, err := confloader.LoadConfig(arg)
					common.Must(err)
					c, err := serial.DecodeJSONConfig(r)
					common.Must(err)
					cf.Override(c, arg)
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
