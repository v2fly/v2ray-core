package commands

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/infra/control/command"
)

// FetchCommand fetches resources
type FetchCommand struct{}

// Name of the command
func (c *FetchCommand) Name() string {
	return "fetch"
}

// Description of the command
func (c *FetchCommand) Description() command.Description {
	return command.Description{
		Short: "Fetch resources",
		Usage: []string{command.ExecutableName + " fetch <url>"},
	}
}

// Execute the command
func (c *FetchCommand) Execute(args []string) error {
	// still parse flags for flag.ErrHelp
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(args) < 1 {
		return newError("empty url")
	}
	content, err := FetchHTTPContent(args[0])
	if err != nil {
		return newError("failed to read HTTP response").Base(err)
	}

	os.Stdout.Write(content)
	return nil
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
	common.Must(command.RegisterCommand(&FetchCommand{}))
}
