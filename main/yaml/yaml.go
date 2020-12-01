package yaml

//go:generate go run v2ray.com/core/common/errors/errorgen

import (
	"fmt"
	"io"
	"os"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/main/confloader"
)

func init() {
	common.Must(core.RegisterConfigLoader(&core.ConfigFormat{
		Name:      []string{"YAML"},
		Extension: []string{".yaml", ".yml"},
		Loader: func(input interface{}) (*core.Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				r, err := confloader.LoadExtConfig(append([]string{"-input=yaml", "-output=protobuf"}, v...), os.Stdin)
				if err != nil {
					return nil, fmt.Errorf("failed to execute v2ctl to convert config file: %s", err)
				}
				return core.LoadConfig("protobuf", "", r)
			case io.Reader:
				r, err := confloader.LoadExtConfig([]string{"--input=yaml", "-output=protobuf", "stdin:"}, os.Stdin)
				if err != nil {
					return nil, fmt.Errorf("failed to execute v2ctl to convert config file: %s", err)
				}
				return core.LoadConfig("protobuf", "", r)
			default:
				return nil, fmt.Errorf("unknown type")
			}
		},
	}))
}
