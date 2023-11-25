package engineering

import (
	"flag"
	"io"
	"os"

	"github.com/vincent-petithory/dataurl"

	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdEncodeDataURLContentType *string

var cmdEncodeDataURL = &base.Command{
	UsageLine: "{{.Exec}} engineering encodeDataURL",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		cmdEncodeDataURLContentType = fs.String("type", "application/vnd.v2ray.subscription-singular", "")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		cmd.Flag.Parse(args)

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			base.Fatalf("%s", err)
		}
		dataURL := dataurl.New(content, *cmdEncodeDataURLContentType)
		dataURL.WriteTo(os.Stdout)
	},
}
