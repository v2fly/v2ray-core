package command

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

// Description of a command
type Description struct {
	Short string
	Usage []string
}

// Command represents a command
type Command interface {
	Name() string
	Description() Description
	Execute(args []string) error
}

var (
	// ExecutableName is the executable name of current binary
	ExecutableName  = "v2ctl"
	commandRegistry = make(map[string]Command)
)

// RegisterCommand registers a command to registry
func RegisterCommand(cmd Command) error {
	entry := strings.ToLower(cmd.Name())
	if entry == "" {
		return newError("empty command name")
	}
	commandRegistry[entry] = cmd
	return nil
}

// GetCommand get command by name
func GetCommand(name string) Command {
	cmd, found := commandRegistry[name]
	if !found {
		return nil
	}
	return cmd
}

type hiddenCommand interface {
	Hidden() bool
}

// PrintUsage prints a list of usage for all commands
func PrintUsage() {
	for name, cmd := range commandRegistry {
		if _, ok := cmd.(hiddenCommand); ok {
			continue
		}
		fmt.Println("   ", name, "\t\t\t", cmd.Description().Short)
	}
	fmt.Printf("\nUse \"%s <command> -h\" for more information.\n", ExecutableName)
}

// ExecuteCommand executes a command
func ExecuteCommand(cmd Command) {
	if err := cmd.Execute(os.Args[2:]); err != nil {
		hasError := false
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err.Error())
			fmt.Fprintln(os.Stderr)
			hasError = true
		}

		for _, line := range cmd.Description().Usage {
			fmt.Println(line)
		}

		if hasError {
			os.Exit(-1)
		}
	}
}

// CommandsCount returns commands count in the registry
func CommandsCount() int {
	return len(commandRegistry)
}

func init() {
	exec, err := os.Executable()
	if err != nil {
		return
	}
	ExecutableName = path.Base(exec)
}
