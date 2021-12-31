package udp

import (
	gonet "net"

	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type dataHandler func(content []byte, address gonet.Addr)

type copyHandler struct {
	onData []dataHandler
}

type CopyOption func(*copyHandler)

func CopyPacketConn(dst internet.AbstractPacketConnWriter, src internet.AbstractPacketConnReader, options ...CopyOption) error {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	var buffer [2048]byte
	for {
		n, addr, err := src.ReadFrom(buffer[:])
		if err != nil {
			return err
		}

		for _, handler := range handler.onData {
			handler(buffer[:n], addr)
		}

		_, err = dst.WriteTo(buffer[:n], addr)
		if err != nil {
			return err
		}
	}
}

func UpdateActivity(timer signal.ActivityUpdater) CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(content []byte, address gonet.Addr) {
			timer.Update()
		})
	}
}
