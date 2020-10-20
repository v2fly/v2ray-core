package commands

import (
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
		Usage: []string{command.ExecutableName + " uuid"},
	}
}

// Execute the command
func (c *UUIDCommand) Execute([]string) error {
	u := uuid.New()
	fmt.Println(u.String())
	return nil
}

func init() {
	common.Must(command.RegisterCommand(&UUIDCommand{}))
}
