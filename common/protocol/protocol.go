package protocol

import (
	"errors"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

var ErrProtoNeedMoreData = errors.New("protocol matches, but need more data to complete sniffing")
