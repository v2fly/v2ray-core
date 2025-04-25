package drain

import "io"

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

type Drainer interface {
	AcknowledgeReceive(size int)
	Drain(reader io.Reader) error
}
