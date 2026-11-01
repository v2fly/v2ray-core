package transportdefaults

import (
	"fmt"
	"testing"
)

func TestMPTCPReleaseWindow(t *testing.T) {
	for minor := 56; minor <= 61; minor++ {
		for _, patch := range []string{"0", "9", "0-custom"} {
			version := fmt.Sprintf("5.%d.%s", minor, patch)
			if !disableMPTCPForVersion(version) {
				t.Errorf("MPTCP must default to disabled in %s", version)
			}
		}
	}

	for _, tc := range []struct {
		version  string
		disabled bool
	}{
		{"5.55.0", true},
		{"5.61", true},
		{"5.62.0", false},
		{"5.62.1", false},
		{"5.100.0", false},
		{"6.0.0", false},
		{"", false},
		{"5", false},
		{"5.", false},
		{"5.next.0", false},
		{"5.-1.0", false},
	} {
		if got := disableMPTCPForVersion(tc.version); got != tc.disabled {
			t.Errorf("version %q: disabled = %t, want %t", tc.version, got, tc.disabled)
		}
	}
}
