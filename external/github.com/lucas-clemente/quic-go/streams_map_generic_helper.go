package quic

import (
	"github.com/v2fly/v2ray-core/external/github.com/cheekybits/genny/generic"

	"github.com/v2fly/v2ray-core/external/github.com/lucas-clemente/quic-go/internal/protocol"
)

// In the auto-generated streams maps, we need to be able to close the streams.
// Therefore, extend the generic.Type with the stream close method.
// This definition must be in a file that Genny doesn't process.
type item interface {
	generic.Type
	closeForShutdown(error)
}

const streamTypeGeneric protocol.StreamType = protocol.StreamTypeUni
