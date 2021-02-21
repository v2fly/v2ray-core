package all

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"github.com/v2fly/v2ray-core/v4/main/commands/helpers"
	"gopkg.in/yaml.v2"
)

var cmdConvert = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} convert [c1.json] [<url>.json] [dir1] ...",
	Short:       "convert config files",
	Long: `
Convert config files between different formats. Files are merged 
before convert if multiple assigned.

Arguments:

	-i, -input <format>
		Specify the input format.
		Available values: "auto", "json", "toml", "yaml"
		Default: "auto"

	-o, -output <format>
		Specify the output format
		Available values: "json", "toml", "yaml", "protobuf" / "pb"
		Default: "json"

	-r
		Load folders recursively.

Examples:

	{{.Exec}} {{.LongName}} -output=protobuf config.json           (1)
	{{.Exec}} {{.LongName}} -input=toml config.toml                (2)
	{{.Exec}} {{.LongName}} "path/to/dir"                          (3)
	{{.Exec}} {{.LongName}} -i yaml -o protobuf c1.yaml <url>.yaml (4)

(1) Convert json to protobuf
(2) Convert toml to json
(3) Merge json files in dir
(4) Merge yaml files and convert to protobuf

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

func setConfArgs(cmd *base.Command) {
	cmd.Flag.StringVar(&inputFormat, "input", core.FormatAuto, "")
	cmd.Flag.StringVar(&inputFormat, "i", core.FormatAuto, "")
	cmd.Flag.StringVar(&outputFormat, "output", "json", "")
	cmd.Flag.StringVar(&outputFormat, "o", "json", "")
	cmd.Flag.BoolVar(&confDirRecursively, "r", false, "")
}
func executeConvert(cmd *base.Command, args []string) {
	setConfArgs(cmd)
	cmd.Flag.Parse(args)
	inputFormat = strings.ToLower(inputFormat)
	outputFormat = strings.ToLower(outputFormat)

	m, err := helpers.LoadConfigToMap(cmd.Flag.Args(), inputFormat, confDirRecursively)
	if err != nil {
		base.Fatalf(err.Error())
	}
	err = merge.ApplyRules(m)
	if err != nil {
		base.Fatalf(err.Error())
	}

	var out []byte
	switch outputFormat {
	case core.FormatJSON:
		out, err = json.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case core.FormatTOML:
		out, err = toml.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case core.FormatYAML:
		out, err = yaml.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case core.FormatProtobuf, core.FormatProtobufShort:
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
		base.Errorf("invalid output format: %s", outputFormat)
		base.Errorf("Run '%s help %s' for details.", base.CommandEnv.Exec, cmd.LongName())
		base.Exit()
	}

	if _, err := os.Stdout.Write(out); err != nil {
		base.Fatalf("failed to write stdout: %s", err)
	}
}
