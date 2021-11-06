package shadowsocks

import (
	"github.com/v2fly/v2ray-core/v4/common"
)

var (
	pluginLoader func(plugin string) SIP003Plugin
	plugins      map[string]func() SIP003Plugin
)

func init() {
	plugins = make(map[string]func() SIP003Plugin)
}

func SetPluginLoader(creator func(plugin string) SIP003Plugin) {
	pluginLoader = creator
}

func RegisterPlugin(name string, creator func() SIP003Plugin) {
	plugins[name] = creator
}

type SIP003Plugin interface {
	Init(localHost string, localPort string, remoteHost string, remotePort string, pluginOpts string, pluginArgs []string, account *MemoryAccount) error
	common.Closable
}
