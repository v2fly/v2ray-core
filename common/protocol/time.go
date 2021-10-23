package protocol

import (
	"github.com/v2fly/v2ray-core/v4/app/ntp/ntptime"
	"github.com/v2fly/v2ray-core/v4/common/dice"
)

type Timestamp int64

type TimestampGenerator func() Timestamp

func NowTime() Timestamp {
	return Timestamp(ntptime.Now().Unix())
}

func NewTimestampGenerator(base Timestamp, delta int) TimestampGenerator {
	return func() Timestamp {
		rangeInDelta := dice.Roll(delta*2) - delta
		return base + Timestamp(rangeInDelta)
	}
}
