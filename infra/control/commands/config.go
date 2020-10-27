package commands

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/common"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/infra/control/command"
)

var ctllog = log.New(os.Stderr, "v2ctl> ", 0)

// ConfigCommand converts json to pb
type ConfigCommand struct{}

// Name of the command
func (c *ConfigCommand) Name() string {
	return "config"
}

// Description of the command
func (c *ConfigCommand) Description() command.Description {
	return command.Description{
		Short: "Convert multiple json config to protobuff",
		Usage: []string{
			fmt.Sprintf("  %s %s config.json c1.json c2.json <url>.json", command.ExecutableName, c.Name()),
		},
	}
}

// Execute the command
func (c *ConfigCommand) Execute(args []string) error {
	// still parse flags for flag.ErrHelp
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(args) < 1 {
		return newError("empty config list")
	}

	conf := &conf.Config{}
	for _, arg := range args {
		ctllog.Println("Read config: ", arg)
		r, err := c.LoadArg(arg)
		common.Must(err)
		c, err := serial.DecodeJSONConfig(r)
		if err != nil {
			ctllog.Fatalln(err)
		}
		conf.Override(c, arg)
	}

	pbConfig, err := conf.Build()
	if err != nil {
		return err
	}

	bytesConfig, err := proto.Marshal(pbConfig)
	if err != nil {
		return newError("failed to marshal proto config").Base(err)
	}

	if _, err := os.Stdout.Write(bytesConfig); err != nil {
		return newError("failed to write proto config").Base(err)
	}

	return nil
}

// LoadArg loads one arg, maybe an remote url, or local file path
func (c *ConfigCommand) LoadArg(arg string) (out io.Reader, err error) {
	var data []byte
	switch {
	case strings.HasPrefix(arg, "http://"), strings.HasPrefix(arg, "https://"):
		data, err = FetchHTTPContent(arg)

	case arg == "stdin:":
		data, err = ioutil.ReadAll(os.Stdin)

	default:
		data, err = ioutil.ReadFile(arg)
	}

	if err != nil {
		return
	}
	out = bytes.NewBuffer(data)
	return
}

func init() {
	common.Must(command.RegisterCommand(&ConfigCommand{}))
}
