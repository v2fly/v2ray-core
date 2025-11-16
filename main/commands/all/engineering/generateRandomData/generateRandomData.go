package generateRandomData

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

var length *int

var cmdGenerateRandomData = &base.Command{
	UsageLine: "{{.Exec}} engineering generate-random-data",
	Short:     "generate random data and output as base64",
	Long: `
Generate random data of specified length and output as base64 encoded string.

Usage:
	{{.Exec}} engineering generate-random-data -length <bytes>

Options:
	-length <bytes>
		The number of random bytes to generate (required)

Example:
	{{.Exec}} engineering generate-random-data -length 32
`,
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		length = fs.Int("length", 0, "number of random bytes to generate")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		err := cmd.Flag.Parse(args)
		if err != nil {
			base.Fatalf("failed to parse flags: %v", err)
		}

		if *length <= 0 {
			base.Fatalf("length must be a positive integer, got: %d", *length)
		}

		// Generate random data
		randomData := make([]byte, *length)
		_, err = rand.Read(randomData)
		if err != nil {
			base.Fatalf("failed to generate random data: %v", err)
		}

		// Encode to base64
		encoded := base64.StdEncoding.EncodeToString(randomData)

		// Output the result
		fmt.Println(encoded)
	},
}

func init() {
	engineering.AddCommand(cmdGenerateRandomData)
}
