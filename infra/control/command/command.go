package command

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
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
	ExecutableName     = "v2ctl"
	commandRegistry    = make(map[string]Command)
	commandListSorted  []string
	commandColumnWidth = 12 // here set the minimal width of command column
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
	if commandListSorted == nil {
		commandListSorted = make([]string, 0)
		for name := range commandRegistry {
			commandListSorted = append(commandListSorted, name)
			len := len(name)
			if commandColumnWidth < len {
				commandColumnWidth = len
			}
		}
		sort.Strings(commandListSorted)
	}
	fmt.Println(ExecutableName, "<command>")
	fmt.Println("Available commands:")
	for _, name := range commandListSorted {
		cmd := commandRegistry[name]
		if _, ok := cmd.(hiddenCommand); ok {
			continue
		}
		nameCol := name + strings.Repeat(" ", commandColumnWidth-len(name))
		fmt.Println("   ", nameCol, cmd.Description().Short)
	}
	fmt.Printf("\nUse \"%s help <command>\" for more information.\n", ExecutableName)
}

// Execute executes a command with args
func Execute(cmd Command, args []string) {
	if err := cmd.Execute(args); err != nil {
		hasError := false
		if err == flag.ErrHelp {
			for _, line := range cmd.Description().Usage {
				fmt.Println(line)
			}
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
			hasError = true
		}
		if hasError {
			os.Exit(-1)
		}
	}
}

func init() {
	exec, err := os.Executable()
	if err != nil {
		return
	}
	ExecutableName = path.Base(exec)
}
