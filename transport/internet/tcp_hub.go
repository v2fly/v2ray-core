package internet

import (
	"context"

	"github.com/ghxhy/v2ray-core/v5/common/net"
)

var transportListenerCache = make(map[string]ListenFunc)

func RegisterTransportListener(protocol string, listener ListenFunc) error {
	if _, found := transportListenerCache[protocol]; found {
		return newError(protocol, " listener already registered.").AtError()
	}
	transportListenerCache[protocol] = listener
	return nil
}

type ConnHandler func(Connection)

type ListenFunc func(ctx context.Context, address net.Address, port net.Port, settings *MemoryStreamConfig, handler ConnHandler) (Listener, error)

type Listener interface {
	Close() error
	Addr() net.Addr
}

// ListenUnix is the UDS version of ListenTCP
func ListenUnix(ctx context.Context, address net.Address, settings *MemoryStreamConfig, handler ConnHandler) (Listener, error) {
	if settings == nil {
		s, err := ToMemoryStreamConfig(nil)
		if err != nil {
			return nil, newError("failed to create default unix stream settings").Base(err)
		}
		settings = s
	}

	protocol := settings.ProtocolName

	if originalProtocolName := getOriginalMessageName(settings); originalProtocolName != "" {
		protocol = originalProtocolName
	}

	listenFunc := transportListenerCache[protocol]
	if listenFunc == nil {
		return nil, newError(protocol, " unix istener not registered.").AtError()
	}
	listener, err := listenFunc(ctx, address, net.Port(0), settings, handler)
	if err != nil {
		return nil, newError("failed to listen on unix address: ", address).Base(err)
	}
	return listener, nil
}

func ListenTCP(ctx context.Context, address net.Address, port net.Port, settings *MemoryStreamConfig, handler ConnHandler) (Listener, error) {
	if settings == nil {
		s, err := ToMemoryStreamConfig(nil)
		if err != nil {
			return nil, newError("failed to create default stream settings").Base(err)
		}
		settings = s
	}

	if address.Family().IsDomain() && address.Domain() == "localhost" {
		address = net.LocalHostIP
	}

	if address.Family().IsDomain() {
		return nil, newError("domain address is not allowed for listening: ", address.Domain())
	}

	protocol := settings.ProtocolName

	if originalProtocolName := getOriginalMessageName(settings); originalProtocolName != "" {
		protocol = originalProtocolName
	}

	listenFunc := transportListenerCache[protocol]
	if listenFunc == nil {
		return nil, newError(protocol, " listener not registered.").AtError()
	}
	listener, err := listenFunc(ctx, address, port, settings, handler)
	if err != nil {
		return nil, newError("failed to listen on address: ", address, ":", port).Base(err)
	}
	return listener, nil
}

// ListenSystem listens on a local address for incoming TCP connections.
//
// v2ray:api:beta
func ListenSystem(ctx context.Context, addr net.Addr, sockopt *SocketConfig) (net.Listener, error) {
	return effectiveListener.Listen(ctx, addr, sockopt)
}

// ListenSystemPacket listens on a local address for incoming UDP connections.
//
// v2ray:api:beta
func ListenSystemPacket(ctx context.Context, addr net.Addr, sockopt *SocketConfig) (net.PacketConn, error) {
	return effectiveListener.ListenPacket(ctx, addr, sockopt)
}
