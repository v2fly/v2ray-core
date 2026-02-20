package wgcommon

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func filterDebugData(in string) string {
	lines := strings.Split(in, "\n")
	outLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLeft := strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(trimmedLeft, "private_key="):
			continue
		case strings.HasPrefix(trimmedLeft, "preshared_key="):
			continue
		case strings.HasPrefix(trimmedLeft, "public_key="):
			leading := line[:len(line)-len(trimmedLeft)]
			value := strings.TrimSpace(trimmedLeft[len("public_key="):])
			decoded, err := hex.DecodeString(value)
			if err != nil {
				outLines = append(outLines, line)
				continue
			}
			encoded := base64.StdEncoding.EncodeToString(decoded)
			outLines = append(outLines, leading+"public_key="+encoded)
			continue
		default:
			outLines = append(outLines, line)
		}
	}

	return strings.Join(outLines, "\n")
}
