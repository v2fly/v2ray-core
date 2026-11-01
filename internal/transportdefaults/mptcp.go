// Package transportdefaults holds release-dependent socket defaults shared by
// the core and transport packages.
package transportdefaults

import (
	"strconv"
	"strings"
	"sync/atomic"
)

var disableMPTCP atomic.Bool

// SetVersion initializes transport defaults for the running V2Ray version.
func SetVersion(version string) {
	disableMPTCP.Store(disableMPTCPForVersion(version))
}

// MPTCPDisabled reports whether unspecified MPTCP settings default to disabled.
func MPTCPDisabled() bool {
	return disableMPTCP.Load()
}

func disableMPTCPForVersion(version string) bool {
	// Temporarily disable MPTCP by default for CVE-2026-98432 through v5.61.x
	// (the next six minor releases after v5.55). Restore Go's default at v5.62.0.
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 || parts[0] != "5" {
		return false
	}
	minor, err := strconv.Atoi(parts[1])
	return err == nil && minor >= 0 && minor < 62
}
