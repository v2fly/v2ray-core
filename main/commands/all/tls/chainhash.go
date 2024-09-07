package tls

import (
	"fmt"
	"os"

	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	v2tls "github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

var cmdChainHash = &base.Command{
	UsageLine: "{{.Exec}} tls certChainHash [--cert <cert.pem>]",
	Short:     "Generate certificate chain hash for given certificate bundle",
}

func init() {
	cmdChainHash.Run = executeChainHash // break init loop
}

var certFile = cmdChainHash.Flag.String("cert", "cert.pem", "")

func executeChainHash(cmd *base.Command, args []string) {
	if len(*certFile) == 0 {
		base.Fatalf("cert file not specified")
	}
	certContent, err := os.ReadFile(*certFile)
	if err != nil {
		base.Fatalf("Failed to read cert file: %s", err)
		return
	}

	certChainHashB64 := v2tls.CalculatePEMCertChainSHA256Hash(certContent)
	fmt.Println(certChainHashB64)
}
