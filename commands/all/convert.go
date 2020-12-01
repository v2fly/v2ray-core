package all

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
	"v2ray.com/core/commands/base"
	"v2ray.com/core/infra/conf/serial"
)

var cmdConvert = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} convert [-format=json] [-r] [c1.json] [<url>.json] [dir1] ...",
	Short:       "Convert multiple json config to protobuf",
	Long: `
Merge and convert V2Ray config files.

Arguments:

	-i, -input
		Input format, options: "json", "yaml". 
		Default "json"

	-o, -output
		Output format, options: "json", "yaml", "protobuf"/"pb".
		Default "json"

	-r
		Load confdir recursively.

Examples:

	{{.Exec}} {{.LongName}} -output=protobuf config.json         (1)
	{{.Exec}} {{.LongName}} -output=yaml config.json             (2)
	{{.Exec}} {{.LongName}} -input=yaml config.yaml              (3)
	{{.Exec}} {{.LongName}} "path/to/dir"                        (4)
	{{.Exec}} {{.LongName}} -i yaml -o yaml c1.yaml <url>.yaml   (5)

(1) Convert json to protobuf
(2) Convert json to yaml
(3) Convert yaml to json
(4) Merge json files in dir into one json
(5) Merge yaml files and convert to yaml

Use "{{.Exec}} help config-merge" for more information about merge.
`,
}

func init() {
	cmdConvert.Run = executeConvert // break init loop
}

var (
	inputFormat        string
	outputFormat       string
	confDirRecursively bool
)
var formatExtensions = map[string][]string{
	"json": {".json", ".jsonc"},
	"yaml": {".yaml", ".yml"},
}

func setConfArgs(cmd *base.Command) {
	cmd.Flag.StringVar(&inputFormat, "input", "json", "")
	cmd.Flag.StringVar(&inputFormat, "i", "json", "")
	cmd.Flag.StringVar(&outputFormat, "output", "json", "")
	cmd.Flag.StringVar(&outputFormat, "o", "json", "")
	cmd.Flag.BoolVar(&confDirRecursively, "r", true, "")
}
func executeConvert(cmd *base.Command, args []string) {
	setConfArgs(cmd)
	cmd.Flag.Parse(args)
	unnamed := cmd.Flag.Args()
	inputFormat = strings.ToLower(inputFormat)
	outputFormat = strings.ToLower(outputFormat)

	files := resolveFolderToFiles(unnamed, formatExtensions[inputFormat], confDirRecursively)
	if len(files) == 0 {
		base.Fatalf("empty config list")
	}
	m := mergeConvertToMap(files, inputFormat)

	var (
		out []byte
		err error
	)
	switch outputFormat {
	case "json":
		out, err = json.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case "yaml":
		out, err = yaml.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case "pb", "protobuf":
		data, err := json.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
		r := bytes.NewReader(data)
		cf, err := serial.DecodeJSONConfig(r)
		if err != nil {
			base.Fatalf("failed to decode json: %s", err)
		}
		pbConfig, err := cf.Build()
		if err != nil {
			base.Fatalf(err.Error())
		}
		out, err = proto.Marshal(pbConfig)
		if err != nil {
			base.Fatalf("failed to marshal proto config: %s", err)
		}
	default:
		base.Fatalf("invalid input format: %s", outputFormat)
	}

	if _, err := os.Stdout.Write(out); err != nil {
		base.Fatalf("failed to write proto config: %s", err)
	}
}
