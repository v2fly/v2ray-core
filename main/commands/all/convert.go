package all

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"

	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	v2json "github.com/v2fly/v2ray-core/v4/infra/conf/json"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/infra/conf/serial"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

var cmdConvert = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} convert [c1.json] [<url>.json] [dir1] ...",
	Short:       "Convert config files",
	Long: `
Convert config files between different formats. Files are merged 
before convert if multiple assigned.

Arguments:

	-i, -input <format>
		Specify the input format.
		Available values: "json", "toml", "yaml"
		Default: "json"

	-o, -output <format>
		Specify the output format
		Available values: "json", "toml", "yaml", "protobuf" / "pb"
		Default: "json"

	-r
		Load confdir recursively.

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
var formatExtensions = map[string][]string{
	"json": {".json", ".jsonc"},
	"toml": {".toml"},
	"yaml": {".yaml", ".yml"},
}

func setConfArgs(cmd *base.Command) {
	cmd.Flag.StringVar(&inputFormat, "input", "json", "")
	cmd.Flag.StringVar(&inputFormat, "i", "json", "")
	cmd.Flag.StringVar(&outputFormat, "output", "json", "")
	cmd.Flag.StringVar(&outputFormat, "o", "json", "")
	cmd.Flag.BoolVar(&confDirRecursively, "r", false, "")
}
func executeConvert(cmd *base.Command, args []string) {
	setConfArgs(cmd)
	cmd.Flag.Parse(args)
	unnamed := cmd.Flag.Args()
	inputFormat = strings.ToLower(inputFormat)
	outputFormat = strings.ToLower(outputFormat)

	var (
		files []string
		err   error
	)
	if len(unnamed) == 0 {
		files = []string{"stdin:"}
	} else {
		files, err = resolveFolderToFiles(unnamed, formatExtensions[inputFormat], confDirRecursively)
		if err != nil {
			base.Fatalf(err.Error())
		}
	}
	m := mergeConvertToMap(files, inputFormat)

	var out []byte
	switch outputFormat {
	case "json":
		out, err = json.Marshal(m)
		if err != nil {
			base.Fatalf("failed to marshal json: %s", err)
		}
	case "toml":
		out, err = toml.Marshal(m)
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
		base.Errorf("invalid output format: %s", outputFormat)
		base.Errorf("Run '%s help %s' for details.", base.CommandEnv.Exec, cmd.LongName())
		base.Exit()
	}

	if _, err := os.Stdout.Write(out); err != nil {
		base.Fatalf("failed to write stdout: %s", err)
	}
}

func mergeConvertToMap(files []string, format string) map[string]interface{} {
	var (
		m   map[string]interface{}
		err error
	)
	switch inputFormat {
	case "json":
		m, err = merge.FilesToMap(files)
		if err != nil {
			base.Fatalf("failed to load json: %s", err)
		}
	case "toml":
		bs, err := tomlsToJSONs(files)
		if err != nil {
			base.Fatalf("failed to convert toml to json: %s", err)
		}
		m, err = merge.BytesToMap(bs)
		if err != nil {
			base.Fatalf("failed to merge converted json: %s", err)
		}
	case "yaml":
		bs, err := yamlsToJSONs(files)
		if err != nil {
			base.Fatalf("failed to convert yaml to json: %s", err)
		}
		m, err = merge.BytesToMap(bs)
		if err != nil {
			base.Fatalf("failed to merge converted json: %s", err)
		}
	default:
		base.Errorf("invalid input format: %s", format)
		base.Errorf("Run '%s help %s' for details.", base.CommandEnv.Exec, cmdConvert.LongName())
		base.Exit()
	}
	return m
}

func yamlsToJSONs(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := v2json.FromYAML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}

func tomlsToJSONs(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := v2json.FromTOML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}
