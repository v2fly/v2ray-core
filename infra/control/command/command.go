package command

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type Description struct {
	Short string
	Usage []string
}

type Command interface {
	Name() string
	Description() Description
	Execute(args []string) error
}

var (
	ExecutableName  = "v2ctl"
	commandRegistry = make(map[string]Command)
)

func RegisterCommand(cmd Command) error {
	entry := strings.ToLower(cmd.Name())
	if entry == "" {
		return newError("empty command name")
	}
	commandRegistry[entry] = cmd
	return nil
}

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

func PrintUsage() {
	for name, cmd := range commandRegistry {
		if _, ok := cmd.(hiddenCommand); ok {
			continue
		}
		fmt.Println("   ", name, "\t\t\t", cmd.Description().Short)
	}
	fmt.Printf("\nUse \"%s <command> -h\" for more information.\n", ExecutableName)
}

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
