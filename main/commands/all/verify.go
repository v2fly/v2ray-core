package all

import (
	"os"

	"github.com/v2fly/VSign/signerVerify"

	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var cmdVerify = &base.Command{
	UsageLine: "{{.Exec}} verify [--sig=sig-file] file",
	Short:     "verify if a binary is officially signed",
	Long: `
Verify if a binary is officially signed.

Arguments:

	-sig <signature_file>
		The path to the signature file
`,
}

func init() {
	cmdVerify.Run = executeVerify // break init loop
}

var verifySigFile = cmdVerify.Flag.String("sig", "", "Path to the signature file")

func executeVerify(cmd *base.Command, args []string) {
	target := cmdVerify.Flag.Arg(0)
	if target == "" {
		base.Fatalf("empty file path.")
	}

	if *verifySigFile == "" {
		base.Fatalf("empty signature path.")
	}

	sigReader, err := os.Open(os.ExpandEnv(*verifySigFile))
	if err != nil {
		base.Fatalf("failed to open file %s: %s ", *verifySigFile, err)
	}

	files := cmdVerify.Flag.Args()

	err = signerVerify.OutputAndJudge(signerVerify.CheckSignaturesV2Fly(sigReader, files))
	if err != nil {
		base.Fatalf("file is not officially signed by V2Ray: %s", err)
	}
}
