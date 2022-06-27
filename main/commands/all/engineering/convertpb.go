package engineering

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var (
	configFiles          cmdarg.Arg
	configDirs           cmdarg.Arg
	configFormat         *string
	configDirRecursively *bool
)

func setConfigFlags(cmd *base.Command) {
	configFormat = cmd.Flag.String("format", core.FormatAuto, "")
	configDirRecursively = cmd.Flag.Bool("r", false, "")

	cmd.Flag.Var(&configFiles, "config", "")
	cmd.Flag.Var(&configFiles, "c", "")
	cmd.Flag.Var(&configDirs, "confdir", "")
	cmd.Flag.Var(&configDirs, "d", "")
}

var cmdConvertPb = &base.Command{
	UsageLine:   "{{.Exec}} engineering convertpb [-c config.json] [-d dir]",
	CustomFlags: true,
	Run: func(cmd *base.Command, args []string) {
		setConfigFlags(cmd)
		cmd.Flag.Parse(args)
		config, err := core.LoadConfig(*configFormat, configFiles)
		if err != nil {
			if len(configFiles) == 0 {
				base.Fatalf("%s", newError("failed to load config").Base(err))
				return
			}
			base.Fatalf("%s", newError(fmt.Sprintf("failed to load config: %s", configFiles)).Base(err))
			return
		}
		bytew, err := proto.Marshal(config)
		if err != nil {
			base.Fatalf("%s", newError("failed to marshal config").Base(err))
			return
		}
		io.Copy(os.Stdout, bytes.NewReader(bytew))
	},
}
