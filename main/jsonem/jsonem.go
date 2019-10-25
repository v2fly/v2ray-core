package jsonem

import (
	"v2ray.com/core/v4"
	"v2ray.com/core/v4/common"
	"v2ray.com/core/v4/infra/conf/serial"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      "JSON",
		Extension: []string{"json"},
		Loader:    serial.LoadJSONConfig,
	}))
}
