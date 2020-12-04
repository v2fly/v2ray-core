package all

import (
	"fmt"

	"v2ray.com/core/common/uuid"
	"v2ray.com/core/main/commands/base"
)

var cmdUUID = &base.Command{
	UsageLine: "{{.Exec}} uuid",
	Short:     "Generate new UUIDs",
	Long: `Generate new UUIDs.
`,
	Run: executeUUID,
}

func executeUUID(cmd *base.Command, args []string) {
	u := uuid.New()
	fmt.Println(u.String())
}
