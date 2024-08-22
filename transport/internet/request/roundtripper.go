package request

import (
	"context"
	"io"

	"github.com/v2fly/v2ray-core/v5/common"
)

type RoundTripperClient interface {
	Tripper
	TransportClientAssemblyReceiver
}

type RoundTripperServer interface {
	common.Runnable
	TransportServerAssemblyReceiver
}

type Tripper interface {
	RoundTrip(ctx context.Context, req Request, opts ...RoundTripperOption) (resp Response, err error)
}

type TripperReceiver interface {
	OnRoundTrip(ctx context.Context, req Request, opts ...RoundTripperOption) (resp Response, err error)
}

type RoundTripperOption interface {
	RoundTripperOption()
}

type Request struct {
	Data          []byte
	ConnectionTag []byte
}

type Response struct {
	Data []byte
}

type OptionSupportsStreamingResponse interface {
	RoundTripperOption
	GetResponseWriter() io.Writer
}

type OptionSupportsStreamingResponseExtensionFlusher interface {
	Flush()
}
