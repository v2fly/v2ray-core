package all

import (
	"v2ray.com/core/commands/base"
)

var docMerge = &base.Command{
	UsageLine: "{{.Exec}} json-merge",
	Short:     "json merge logic",
	Long: `
Merging of JSON configs is applied in following commands:

	{{.Exec}} run -c c1.json -c c2.json ...
	{{.Exec}} merge c1.json https://url.to/c2.json ...
	{{.Exec}} convert c1.json dir1 ...

Suppose we have 2 JSON files,

The 1st one:

	{
	  "log": {"access": "some_value", "loglevel": "debug"},
	  "inbounds": [{"tag": "in-1"}],
	  "outbounds": [{"priority": 100, "tag": "out-1"}],
	  "routing": {"rules": [{"inboundTag":["in-1"],"outboundTag":"out-1"}]}
	}

The 2nd one:

	{
	  "log": {"loglevel": "error"},
	  "inbounds": [{"tag": "in-2"}],
	  "outbounds": [{"priority": -100, "tag": "out-2"}],
	  "routing": {"rules": [{"inboundTag":["in-2"],"outboundTag":"out-2"}]}
	}

Output:

	{
	  "log": {"access": "some_value", "loglevel": "error"},
	  "inbounds": [{"tag": "in-1"}, {"tag": "in-2"}],
	  "outbounds": [
		{"tag": "out-2"}, // note the order is affected by priority
		{"tag": "out-1"}
	  ],
	  "routing": {"rules": [
		{"inboundTag":["in-1"],"outboundTag":"out-1"}
		{"inboundTag":["in-2"],"outboundTag":"out-2"}
	  ]}
	}

Explained: 

- Simple values (string, number, boolean) are override, all others are merged
- Add "priority" property to array elements will help sort the array
`,
}
