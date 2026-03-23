package dtls

import (
	"github.com/pion/dtls/v3"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const protocolName = "dtls"

type ClientIdentityProvider interface {
	ClientIdentity() []byte
}

type connectionStateProvider interface {
	ConnectionState() (dtls.State, bool)
}

func ClientIdentity(conn internet.Connection) []byte {
	if provider, ok := conn.(ClientIdentityProvider); ok {
		return append([]byte(nil), provider.ClientIdentity()...)
	}
	if provider, ok := conn.(connectionStateProvider); ok {
		state, ready := provider.ConnectionState()
		if !ready {
			return nil
		}
		return append([]byte(nil), state.IdentityHint...)
	}
	return nil
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
