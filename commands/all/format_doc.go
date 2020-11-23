package all

import (
	"v2ray.com/core/commands/base"
)

var docFormat = &base.Command{
	UsageLine: "{{.Exec}} format-loader",
	Short:     "config formats and loading",
	Long: `
{{.Exec}} supports different config format:

	* json (.json, .jsonc)
	  The default loader, multiple config files support.

	* yaml (.yml)
	  The yaml loader (comming soon?), multiple config files support.

	* protobuf / pb (.pb)
	  Single conifg file support. If multiple files assigned, 
	  only the first one is loaded.

If "-format" is not explicitly specified, {{.Exec}} will choose 
a loader by detecting the extension of first config file, or use 
the default loader.

The following explains format loaders' behavior with examples.

Examples:

	{{.Exec}} run -d dir                                  (1)
	{{.Exec}} run -format=protobuf -d dir                 (2)
	{{.Exec}} test -c c1.yml -d dir                       (3)
	{{.Exec}} test -format=pb -c c1.json                  (4)

(1) The default json loader is used, {{.Exec}} will try to load all 
	json files in the "dir".

(2) The protobuf loader is specified, {{.Exec}} will try to find 
	all protobuf files in the "dir", but only the the first 
	.pb file is loaded.

(3) The yaml loader is selected because of the "c1.yml" file, 
	{{.Exec}} will try to load all yaml files in the "dir".

(4) The protobuf loader is specified, {{.Exec}} will load 
	"c1.json" as protobuf, no matter its extension.
`,
}
