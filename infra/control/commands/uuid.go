package commands

import (
	"flag"
	"fmt"

	"v2ray.com/core/common"
	"v2ray.com/core/common/uuid"
	"v2ray.com/core/infra/control/command"
)

// UUIDCommand generates new UUIDs
type UUIDCommand struct{}

// Name of the command
func (c *UUIDCommand) Name() string {
	return "uuid"
}

// Description of the command
func (c *UUIDCommand) Description() command.Description {
	return command.Description{
		Short: "Generate new UUIDs",
		Usage: []string{
			fmt.Sprintf("  %s %s", command.ExecutableName, c.Name()),
		},
	}
}

// Execute the command
func (c *UUIDCommand) Execute(args []string) error {
	// still parse flags for flag.ErrHelp
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	u := uuid.New()
	fmt.Println(u.String())
	return nil
}

func init() {
	common.Must(command.RegisterCommand(&UUIDCommand{}))
}
