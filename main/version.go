package main

import (
	"flag"
	"fmt"

	"v2ray.com/core/common"
	"v2ray.com/core/infra/control/command"
)

type versionCommand struct{}

// Name of the command
func (c *versionCommand) Name() string {
	return "version"
}

// Description of the command
func (c *versionCommand) Description() command.Description {
	return command.Description{
		Short: "print V2Ray version",
		Usage: []string{
			fmt.Sprintf("  %s %s", command.ExecutableName, c.Name()),
		},
	}
}

// Execute the command
func (c *versionCommand) Execute(args []string) error {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	printVersion()
	return nil
}

func init() {
	common.Must(command.RegisterCommand(&versionCommand{}))
}
