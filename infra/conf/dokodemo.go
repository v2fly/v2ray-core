package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v4/proxy/dokodemo"
)

type DokodemoConfig struct {
	Host         *cfgcommon.Address     `json:"address"`
	PortValue    uint16                 `json:"port"`
	NetworkList  *cfgcommon.NetworkList `json:"network"`
	TimeoutValue uint32                 `json:"timeout"`
	Redirect     bool                   `json:"followRedirect"`
	UserLevel    uint32                 `json:"userLevel"`
}

func (v *DokodemoConfig) Build() (proto.Message, error) {
	config := new(dokodemo.Config)
	if v.Host != nil {
		config.Address = v.Host.Build()
	}
	config.Port = uint32(v.PortValue)
	config.Networks = v.NetworkList.Build()
	config.Timeout = v.TimeoutValue
	config.FollowRedirect = v.Redirect
	config.UserLevel = v.UserLevel
	return config, nil
}
