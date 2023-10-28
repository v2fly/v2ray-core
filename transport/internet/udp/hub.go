package udp

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type HubOption func(h *Hub)

func HubCapacity(capacity int) HubOption {
	return func(h *Hub) {
		h.capacity = capacity
	}
}

func HubReceiveOriginalDestination(r bool) HubOption {
	return func(h *Hub) {
		h.recvOrigDest = r
	}
}

type Hub struct {
	conn         *net.UDPConn
	connPacket   net.PacketConn
	cache        chan *udp.Packet
	capacity     int
	recvOrigDest bool
}

func ListenUDP(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, options ...HubOption) (*Hub, error) {
	hub := &Hub{
		capacity:     256,
		recvOrigDest: false,
	}
	for _, opt := range options {
		opt(hub)
	}

	var sockopt *internet.SocketConfig
	if streamSettings != nil {
		sockopt = streamSettings.SocketSettings
	}
	if sockopt != nil && sockopt.ReceiveOriginalDestAddress {
		hub.recvOrigDest = true
	}
	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	listener := transportEnvironment.Listener()
	udpConn, err := listener.ListenPacket(ctx, &net.UDPAddr{
		IP:   address.IP(),
		Port: int(port),
	}, sockopt)
	if err != nil {
		return nil, err
	}
	newError("listening UDP on ", address, ":", port).WriteToLog()
	if udpConnDirect, ok := udpConn.(*net.UDPConn); ok {
		hub.conn = udpConnDirect
	} else {
		hub.connPacket = udpConn
	}

	hub.cache = make(chan *udp.Packet, hub.capacity)

	go hub.start()
	return hub, nil
}

// Close implements net.Listener.
func (h *Hub) Close() error {
	if h.connPacket != nil {
		h.connPacket.Close()
		return nil
	}
	h.conn.Close()
	return nil
}

func (h *Hub) WriteTo(payload []byte, dest net.Destination) (int, error) {
	if h.connPacket != nil {
		return h.connPacket.WriteTo(payload, &net.UDPAddr{
			IP:   dest.Address.IP(),
			Port: int(dest.Port),
		})
	}
	return h.conn.WriteToUDP(payload, &net.UDPAddr{
		IP:   dest.Address.IP(),
		Port: int(dest.Port),
	})
}

func (h *Hub) start() {
	c := h.cache
	defer close(c)

	oobBytes := make([]byte, 256)

	for {
		buffer := buf.New()
		if h.conn != nil {
			var noob int
			var addr *net.UDPAddr
			rawBytes := buffer.Extend(buf.Size)
			n, noob, _, addr, err := ReadUDPMsg(h.conn, rawBytes, oobBytes)
			if err != nil {
				newError("failed to read UDP msg").Base(err).WriteToLog()
				buffer.Release()
				break
			}
			buffer.Resize(0, int32(n))

			if buffer.IsEmpty() {
				buffer.Release()
				continue
			}

			payload := &udp.Packet{
				Payload: buffer,
				Source:  net.UDPDestination(net.IPAddress(addr.IP), net.Port(addr.Port)),
			}
			if h.recvOrigDest && noob > 0 {
				payload.Target = RetrieveOriginalDest(oobBytes[:noob])
				if payload.Target.IsValid() {
					newError("UDP original destination: ", payload.Target).AtDebug().WriteToLog()
				} else {
					newError("failed to read UDP original destination").WriteToLog()
				}
			}

			select {
			case c <- payload:
			default:
				buffer.Release()
				payload.Payload = nil
			}
		} else {
			rawBytes := buffer.Extend(buf.Size)
			n, addr, err := h.connPacket.ReadFrom(rawBytes)
			if err != nil {
				newError("failed to read UDP msg").Base(err).WriteToLog()
				buffer.Release()
				break
			}
			buffer.Resize(0, int32(n))

			if buffer.IsEmpty() {
				buffer.Release()
				continue
			}

			payload := &udp.Packet{
				Payload: buffer,
				Source:  net.DestinationFromAddr(addr),
			}
			select {
			case c <- payload:
			default:
				buffer.Release()
				payload.Payload = nil
			}
		}
	}
}

// Addr implements net.Listener.
func (h *Hub) Addr() net.Addr {
	if h.conn == nil {
		return h.connPacket.LocalAddr()
	}
	return h.conn.LocalAddr()
}

func (h *Hub) Receive() <-chan *udp.Packet {
	return h.cache
}
