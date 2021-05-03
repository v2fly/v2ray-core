package control

import (
	"fmt"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

type UUIDCommand struct{}

func (c *UUIDCommand) Name() string {
	return "uuid"
}

func (c *UUIDCommand) Description() Description {
	return Description{
		Short: "Generate new UUID",
		Usage: []string{"v2ctl uuid"},
	}
}

func (c *UUIDCommand) Execute([]string) error {
	u := uuid.New()
	fmt.Println(u.String())
	return nil
}

func init() {
	common.Must(RegisterCommand(&UUIDCommand{}))
}
