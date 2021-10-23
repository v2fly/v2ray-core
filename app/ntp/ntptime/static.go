package ntptime

import (
	"time"

	"github.com/v2fly/v2ray-core/v4/features/extension"
)

var Instance extension.NTPClient

func Now() time.Time {
	if Instance == nil {
		return time.Now()
	}
	return Instance.FixedNow()
}
