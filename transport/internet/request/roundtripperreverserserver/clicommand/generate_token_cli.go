package clicommand

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripperreverserserver"
)

var (
	genPassword      *string
	genPasswordShort *string
	genUser          *string
	genOut           *string
)

var cmdGenerateToken = &base.Command{
	UsageLine: "{{.Exec}} engineering request-rtt-reverser-gen-token",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		// user-facing flag name: access passphrase
		genPassword = fs.String("access-passphrase", "", "access passphrase used to derive token key (required)")
		// short alias for convenience
		genPasswordShort = fs.String("p", "", "access passphrase (shorthand)")
		genUser = fs.String("u", "", "userID (required, int64, positive)")
		genOut = fs.String("o", "", "write base64 tokens to file (optional)")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		if err := cmd.Flag.Parse(args); err != nil {
			base.Fatalf("failed to parse flags: %v", err)
		}

		// require password (either long form or short alias)
		var passphrase string
		if genPasswordShort != nil && *genPasswordShort != "" {
			passphrase = *genPasswordShort
		} else if genPassword != nil && *genPassword != "" {
			passphrase = *genPassword
		} else {
			base.Fatalf("-access-passphrase (or -p) is required")
		}

		if *genUser == "" {
			base.Fatalf("-u userID is required")
		}
		id, err := strconv.ParseInt(*genUser, 10, 64)
		if err != nil {
			base.Fatalf("invalid userID %q: %v", *genUser, err)
		}

		// require positive userID
		if id <= 0 {
			base.Fatalf("userID must be a positive integer")
		}

		checker, err := roundtripperreverserserver.NewPasswordAccessChecker(passphrase)
		if err != nil {
			base.Fatalf("failed to create PasswordAccessChecker: %v", err)
		}

		// generate private (positive) and public (negative) tokens
		privToken, err := checker.GenerateToken(id)
		if err != nil {
			base.Fatalf("GenerateToken (private) failed: %v", err)
		}
		pubToken, err := checker.GenerateToken(-id)
		if err != nil {
			base.Fatalf("GenerateToken (public) failed: %v", err)
		}

		b64Priv := base64.StdEncoding.EncodeToString(privToken)
		b64Pub := base64.StdEncoding.EncodeToString(pubToken)

		if *genOut != "" {
			content := fmt.Sprintf("private: %s\npublic: %s\n", b64Priv, b64Pub)
			if err := os.WriteFile(*genOut, []byte(content), 0o600); err != nil {
				base.Fatalf("failed to write tokens to file %q: %v", *genOut, err)
			}
			if _, err := fmt.Fprintf(os.Stdout, "wrote base64 tokens to %s\n", *genOut); err != nil {
				base.Fatalf("failed to write token message: %v", err)
			}
			return
		}

		// print both tokens to stdout
		fmt.Printf("private: %s\npublic: %s\n", b64Priv, b64Pub)
	},
}

func init() {
	engineering.AddCommand(cmdGenerateToken)
}
