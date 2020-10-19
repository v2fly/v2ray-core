package commands

import (
	"fmt"

	"v2ray.com/core/common"
	"v2ray.com/core/common/uuid"
	"v2ray.com/core/infra/control/command"
)

type UUIDCommand struct{}

func (c *UUIDCommand) Name() string {
	return "uuid"
}

func (c *UUIDCommand) Description() command.Description {
	return command.Description{
		Short: "Generate new UUIDs",
		Usage: []string{command.ExecutableName + " uuid"},
	}
}

func (c *UUIDCommand) Execute([]string) error {
	u := uuid.New()
	fmt.Println(u.String())
	return nil
}

func init() {
	common.Must(command.RegisterCommand(&UUIDCommand{}))
}
