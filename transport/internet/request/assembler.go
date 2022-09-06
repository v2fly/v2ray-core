package request

import (
	"io"
)

type SessionAssemblerClient interface {
	SessionCreator
	TransportClientAssemblyReceiver
}
type SessionAssemblerServer interface {
	TripperReceiver
	TransportServerAssemblyReceiver
}

type SessionOption interface {
	RoundTripperOption()
}

type Session interface {
	io.ReadWriteCloser
}
