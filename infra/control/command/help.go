package command

import (
	"fmt"

	"v2ray.com/core/common"
)

// HelpCommand shows usage info of commands, it's a built-in command
type HelpCommand struct{}

// Name of the command
func (c *HelpCommand) Name() string {
	return "help"
}

// Description of the command
func (c *HelpCommand) Description() Description {
	return Description{
		Short: "Show help",
		Usage: []string{ExecutableName + " help"},
	}
}

// Execute the command
func (c *HelpCommand) Execute(args []string) error {
	if len(args) == 0 || args[0] == c.Name() {
		PrintUsage()
		return nil
	}
	cmd := GetCommand(args[0])
	if cmd == nil {
		return fmt.Errorf("unknown help topic '%s'. Run '%s help'", args[0], ExecutableName)
	}
	Execute(cmd, []string{"-h"})
	return nil
}

func init() {
	common.Must(RegisterCommand(&HelpCommand{}))
}
