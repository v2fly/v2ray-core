package extension

import (
	"time"

	"github.com/v2fly/v2ray-core/v4/features"
)

type NTPClient interface {
	features.Feature

	FixedNow() time.Time
}

func NTPType() interface{} {
	return (*NTPClient)(nil)
}
