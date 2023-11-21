package engineering

import (
	"bytes"
	"flag"
	"io"
	"os"

	"github.com/v2fly/v2ray-core/v5/app/subscription/entries/nonnative"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdNonNativeLinkExecInputName *string

var cmdNonNativeLinkExecTemplatePath *string

var cmdNonNativeLinkExec = &base.Command{
	UsageLine: "{{.Exec}} engineering nonnativelinkexec",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		cmdNonNativeLinkExecInputName = fs.String("name", "", "")
		cmdNonNativeLinkExecTemplatePath = fs.String("templatePath", "", "path for template directory (WARNING: This will not stop templates from reading file outside this directory)")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		cmd.Flag.Parse(args)

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			base.Fatalf("%s", err)
		}
		flattenedLink := nonnative.ExtractAllValuesFromBytes(content)

		matcher := nonnative.NewDefMatcher()
		if *cmdNonNativeLinkExecTemplatePath != "" {
			osFs := os.DirFS(*cmdNonNativeLinkExecTemplatePath)
			err = matcher.LoadDefinitions(osFs)
			if err != nil {
				base.Fatalf("%s", err)
			}
		} else {
			err = matcher.LoadEmbeddedDefinitions()
			if err != nil {
				base.Fatalf("%s", err)
			}
		}

		spec, err := matcher.ExecuteNamed(flattenedLink, *cmdNonNativeLinkExecInputName)
		if err != nil {
			base.Fatalf("%s", err)
		}
		io.Copy(os.Stdout, bytes.NewReader(spec))
	},
}
