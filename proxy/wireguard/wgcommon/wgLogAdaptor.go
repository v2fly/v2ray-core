package wgcommon

import (
	"fmt"

	"golang.zx2c4.com/wireguard/device"

	"github.com/v2fly/v2ray-core/v5/common/errors"
)

// NewDeviceLoggerAdapter returns a wireguard device.Logger that forwards
// verbose and error logs into the project's error logger using errors.New(...).
// Verbosef logs are recorded as Debug, Errorf logs are recorded as Error.
// machine generated
func NewDeviceLoggerAdapter() *device.Logger {
	l := &device.Logger{}
	l.Verbosef = func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		err := errors.New(msg)
		err.AtDebug().WriteToLog()
	}
	l.Errorf = func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		err := errors.New(msg)
		err.AtError().WriteToLog()
	}
	return l
}
