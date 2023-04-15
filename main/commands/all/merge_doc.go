package all

import (
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var docMerge = &base.Command{
	UsageLine: "{{.Exec}} config-merge",
	Short:     "config merge logic",
	Long: `
Merging of config files is applied in following commands:

	{{.Exec}} run -c c1.json -c c2.json ...
	{{.Exec}} test -c c1.yaml -c c2.yaml ...
	{{.Exec}} convert c1.json dir1 ...
	{{.Exec}} api ado c1.json dir1 ...
	{{.Exec}} api rmi c1.json dir1 ...
	... and more ...

Support of toml and yaml is implemented by converting them to json, 
both merge and load. So we take json as example here.

Suppose we have 2 JSON files,

The 1st one:

	{
	  "log": {"access": "some_value", "loglevel": "debug"},
	  "inbounds": [{"tag": "in-1"}],
	  "outbounds": [{"_priority": 100, "tag": "out-1"}],
	  "routing": {"rules": [
		{"_tag":"default_route","inboundTag":["in-1"],"outboundTag":"out-1"}
	  ]}
	}

The 2nd one:

	{
	  "log": {"loglevel": "error"},
	  "inbounds": [{"tag": "in-2"}],
	  "outbounds": [{"_priority": -100, "tag": "out-2"}],
	  "routing": {"rules": [
		{"inboundTag":["in-2"],"outboundTag":"out-2"},
		{"_tag":"default_route","inboundTag":["in-1.1"],"outboundTag":"out-1.1"}
	  ]}
	}

Output:

	{
	  // loglevel is overwritten
	  "log": {"access": "some_value", "loglevel": "error"},
	  "inbounds": [{"tag": "in-1"}, {"tag": "in-2"}],
	  "outbounds": [
		{"tag": "out-2"}, // note the order is affected by priority
		{"tag": "out-1"}
	  ],
	  "routing": {"rules": [
		// note 3 rules are merged into 2, and outboundTag is overwritten,
		// because 2 of them has same tag
		{"inboundTag":["in-1","in-1.1"],"outboundTag":"out-1.1"}
		{"inboundTag":["in-2"],"outboundTag":"out-2"}
	  ]}
	}

Explained: 

- Simple values (string, number, boolean) are overwritten, others are merged
- Elements with same "tag" (or "_tag") in an array will be merged
- Add "_priority" property to array elements will help sort the array
`,
}
