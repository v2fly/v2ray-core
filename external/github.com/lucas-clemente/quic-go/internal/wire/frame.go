package wire

import (
	"bytes"

	"github.com/v2fly/v2ray-core/external/github.com/lucas-clemente/quic-go/internal/protocol"
)

// A Frame in QUIC
type Frame interface {
	Write(b *bytes.Buffer, version protocol.VersionNumber) error
	Length(version protocol.VersionNumber) protocol.ByteCount
}
