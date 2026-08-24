package base64urlline

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// TestParseLargeDocument makes sure that server specs stay intact when the
// document is larger than the internal scanner buffer, which forces
// bufio.Scanner to reuse (and overwrite) the memory backing previous tokens.
func TestParseLargeDocument(t *testing.T) {
	const lineCount = 6000

	lines := make([]string, 0, lineCount)
	for i := 0; i < lineCount; i++ {
		lines = append(lines, fmt.Sprintf("vmess://server-%06d-%s", i, strings.Repeat("x", 100)))
	}
	document := base64.StdEncoding.EncodeToString([]byte(strings.Join(lines, "\n")))

	container, err := newBase64URLLineParser().ParseSubscriptionContainerDocument([]byte(document))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}
	if len(container.ServerSpecs) != lineCount {
		t.Fatalf("got %v server specs, want %v", len(container.ServerSpecs), lineCount)
	}
	for i, serverSpec := range container.ServerSpecs {
		if string(serverSpec.Content) != lines[i] {
			t.Fatalf("server spec %v is %q, want %q", i, string(serverSpec.Content), lines[i])
		}
	}
}
