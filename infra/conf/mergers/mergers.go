package mergers

import (
	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/infra/conf/json"
)

func init() {
	common.Must(RegisterMergeLoader(makeLoader(
		core.FormatJSON,
		[]string{".json", ".jsonc"},
		nil,
	)))
	common.Must(RegisterMergeLoader(makeLoader(
		core.FormatTOML,
		[]string{".toml"},
		json.FromTOML,
	)))
	common.Must(RegisterMergeLoader(makeLoader(
		core.FormatYAML,
		[]string{".yml", ".yaml"},
		json.FromYAML,
	)))
	common.Must(RegisterMergeLoader(
		&MergeableFormat{
			Name:       core.FormatAuto,
			Extensions: nil,
			Loader:     Merge,
		}),
	)
}
