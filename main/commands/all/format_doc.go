package all

import (
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var docFormat = &base.Command{
	UsageLine: "{{.Exec}} format-loader",
	Short:     "config formats and loading",
	Long: `
{{.Exec}} is equipped with multiple loaders to support different 
config formats:

	* auto
	  The default loader, supports all formats listed below, with 
	  format detecting, and mixed fomats support.

	* json (.json, .jsonc)
	  The json loader, multiple files support, mergeable.

	* toml (.toml)
	  The toml loader, multiple files support, mergeable.

	* yaml (.yml, .yaml)
	  The yaml loader, multiple files support, mergeable.

	* protobuf / pb (.pb)
	  Single file support, unmergeable.


The following explains how format loaders behaves.

Examples:

	{{.Exec}} run -d dir                        (1)
	{{.Exec}} run -c c1.json -c c2.yaml         (2)
	{{.Exec}} run -format=json -d dir           (3)
	{{.Exec}} test -c c1.yml -c c2.pb           (4)
	{{.Exec}} test -format=pb -d dir            (5)
	{{.Exec}} test -format=protobuf -c c1.json  (6)

(1) Load all supported files in the "dir".
(2) JSON and YAML are merged and loaded.
(3) Load all JSON files in the "dir".
(4) Goes error since .pb is not mergeable to others
(5) Works only when single .pb file found, if not, failed due to 
	unmergeable.
(6) Force load "c1.json" as protobuf, no matter its extension.
`,
}
