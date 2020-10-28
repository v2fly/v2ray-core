package commands

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/infra/control/command"
)

var ctllog = log.New(os.Stderr, "v2ctl> ", 0)

// ConfigCommand converts json to pb
type ConfigCommand struct{}

// Name of the command
func (c *ConfigCommand) Name() string {
	return "convert"
}

// Description of the command
func (c *ConfigCommand) Description() command.Description {
	return command.Description{
		Short: "Convert multiple json config to protobuf",
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

// FetchHTTPContent dials https for remote content
func FetchHTTPContent(target string) ([]byte, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, newError("invalid URL: ", target).Base(err)
	}

	if s := strings.ToLower(parsedTarget.Scheme); s != "http" && s != "https" {
		return nil, newError("invalid scheme: ", parsedTarget.Scheme)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    parsedTarget,
		Close:  true,
	})
	if err != nil {
		return nil, newError("failed to dial to ", target).Base(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newError("unexpected HTTP status code: ", resp.StatusCode)
	}

	content, err := buf.ReadAllToBytes(resp.Body)
	if err != nil {
		return nil, newError("failed to read HTTP response").Base(err)
	}

	return content, nil
}

func init() {
	common.Must(command.RegisterCommand(&ConfigCommand{}))
}
