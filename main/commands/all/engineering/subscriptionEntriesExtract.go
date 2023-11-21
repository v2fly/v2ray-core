package engineering

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/sha3"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdSubscriptionEntriesExtractInputName *string

var cmdSubscriptionEntriesExtract = &base.Command{
	UsageLine: "{{.Exec}} engineering subscriptionEntriesExtract",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		cmdSubscriptionEntriesExtractInputName = fs.String("input", "", "")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		cmd.Flag.Parse(args)
		inputReader := os.Stdin
		if *cmdSubscriptionEntriesExtractInputName != "" {
			file, err := os.Open(*cmdSubscriptionEntriesExtractInputName)
			if err != nil {
				base.Fatalf("%s", err)
			}
			inputReader = file
			defer file.Close()
		}
		content, err := io.ReadAll(inputReader)
		if err != nil {
			base.Fatalf("%s", err)
		}
		parsed, err := containers.TryAllParsers(content, "")
		if err != nil {
			base.Fatalf("%s", err)
		}
		zipWriter := zip.NewWriter(os.Stdout)
		{
			writer, err := zipWriter.Create("meta.json")
			if err != nil {
				base.Fatalf("%s", err)
			}
			err = json.NewEncoder(writer).Encode(parsed.Metadata)
			if err != nil {
				base.Fatalf("%s", err)
			}
		}
		for k, entry := range parsed.ServerSpecs {
			hash := sha3.Sum256(entry.Content)
			fileName := fmt.Sprintf("entry_%v_%x", k, hash[:8])
			writer, err := zipWriter.Create(fileName)
			if err != nil {
				base.Fatalf("%s", err)
			}
			_, err = writer.Write(entry.Content)
			if err != nil {
				base.Fatalf("%s", err)
			}
		}
		zipWriter.Close()
	},
}
