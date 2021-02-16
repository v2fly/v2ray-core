package task

import "github.com/v2fly/v2ray-core/v4/common"

// Close returns a func() that closes v.
func Close(v interface{}) func() error {
	return func() error {
		return common.Close(v)
	}
}
