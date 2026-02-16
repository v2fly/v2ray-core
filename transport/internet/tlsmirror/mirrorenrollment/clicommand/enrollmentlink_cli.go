package clicommand

import (
	"flag"
	"io"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	anypb "google.golang.org/protobuf/types/known/anypb"

	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	mirrorenrollment "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment"
)

var (
	inputPath  *string
	outputPath *string
	mode       *string
)

var cmdEnrollmentLink = &base.Command{
	UsageLine: "{{.Exec}} engineering tlsmirror-enrollment-link",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("tlsmirror-enrollment-link", flag.ExitOnError)
		inputPath = fs.String("c", "", "input file path (optional, defaults to stdin)")
		outputPath = fs.String("o", "", "output file path (default stdout)")
		mode = fs.String("mode", "link", "conversion mode: 'link' to convert JSON -> link, 'json' to convert link -> JSON")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		if err := cmd.Flag.Parse(args); err != nil {
			base.Fatalf("failed to parse flags: %v", err)
		}

		var content []byte
		var err error
		if *inputPath == "" {
			// Read from stdin when -c is omitted.
			content, err = io.ReadAll(os.Stdin)
			if err != nil {
				base.Fatalf("failed to read from stdin: %v", err)
			}
		} else {
			fd, err := os.Open(*inputPath)
			if err != nil {
				base.Fatalf("failed to open input file %q: %v", *inputPath, err)
			}
			defer fd.Close()

			content, err = io.ReadAll(fd)
			if err != nil {
				base.Fatalf("failed to read input file %q: %v", *inputPath, err)
			}
		}

		var outBytes []byte
		switch *mode {
		case "link":
			// Expect protobuf JSON for google.protobuf.Any, convert to a data URL link.
			var any anypb.Any
			if err := protojson.Unmarshal(content, &any); err != nil {
				base.Fatalf("failed to unmarshal JSON into google.protobuf.Any: %v", err)
			}
			link, err := mirrorenrollment.LinkFromAny(&any)
			if err != nil {
				base.Fatalf("failed to create link from Any: %v", err)
			}
			outBytes = []byte(link)

		case "json":
			// Expect link (data URL or other supported forms), convert to protobuf JSON.
			link := string(content)
			any, err := mirrorenrollment.AnyFromLink(link)
			if err != nil {
				base.Fatalf("failed to parse link into Any: %v", err)
			}
			b, err := protojson.Marshal(any)
			if err != nil {
				base.Fatalf("failed to marshal Any to JSON: %v", err)
			}
			outBytes = b

		default:
			base.Fatalf("unknown mode: %s", *mode)
		}

		if *outputPath == "" {
			if _, err := os.Stdout.Write(outBytes); err != nil {
				base.Fatalf("failed to write output to stdout: %v", err)
			}
			return
		}

		if err := os.WriteFile(*outputPath, outBytes, 0o644); err != nil {
			base.Fatalf("failed to write output file %q: %v", *outputPath, err)
		}
	},
}

func init() {
	engineering.AddCommand(cmdEnrollmentLink)
}
