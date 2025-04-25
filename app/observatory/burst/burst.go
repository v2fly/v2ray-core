package burst

import (
	"math"
	"time"
)

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

const (
	rttFailed      = time.Duration(math.MaxInt64 - iota)
	rttUntested    // nolint: varcheck
	rttUnqualified // nolint: varcheck
)
