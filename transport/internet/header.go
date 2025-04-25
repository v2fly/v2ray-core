package internet

import (
	"context"
	"net"

	"github.com/ghxhy/v2ray-core/v5/common"
)

type PacketHeader interface {
	Size() int32
	Serialize([]byte)
}

func CreatePacketHeader(config interface{}) (PacketHeader, error) {
	header, err := common.CreateObject(context.Background(), config)
	if err != nil {
		return nil, err
	}
	if h, ok := header.(PacketHeader); ok {
		return h, nil
	}
	return nil, newError("not a packet header")
}

type ConnectionAuthenticator interface {
	Client(net.Conn) net.Conn
	Server(net.Conn) net.Conn
}

func CreateConnectionAuthenticator(config interface{}) (ConnectionAuthenticator, error) {
	auth, err := common.CreateObject(context.Background(), config)
	if err != nil {
		return nil, err
	}
	if a, ok := auth.(ConnectionAuthenticator); ok {
		return a, nil
	}
	return nil, newError("not a ConnectionAuthenticator")
}
