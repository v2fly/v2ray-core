package all

import (
	"fmt"

	"github.com/v2fly/v2ray-core/v4/commands/base"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
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
