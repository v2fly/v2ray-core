package base

// RootCommand is the root command of all commands
var RootCommand *Command

func init() {
	RootCommand = &Command{
		UsageLine: CommandEnv.Exec,
		Long:      "The root command",
	}
}

// RegisterCommand register a command to V2Ray
func RegisterCommand(cmd *Command) {
	RootCommand.Commands = append(RootCommand.Commands, cmd)
}
