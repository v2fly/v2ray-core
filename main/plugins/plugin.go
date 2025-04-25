package plugins

import "github.com/ghxhy/v2ray-core/v5/main/commands/base"

var Plugins []Plugin

type Plugin func(*base.Command) func() error

func RegisterPlugin(plugin Plugin) {
	Plugins = append(Plugins, plugin)
}
