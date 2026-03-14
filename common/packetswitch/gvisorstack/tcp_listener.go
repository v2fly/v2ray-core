package gvisorstack

import (
	"fmt"
	"io"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tagged"
)

func (w *WrappedStack) ApplyListeners() error {
	for _, listener := range w.config.TcpListener {
		listenerIP := listener.Address.Ip
		listenerPort := listener.GetPort()
		listenerTag := listener.Tag

		addr := tcpip.FullAddress{
			Addr: tcpip.AddrFromSlice(listenerIP),
			Port: uint16(listenerPort),
		}

		netProto := ipv4.ProtocolNumber
		if len(listenerIP) == net.IPv6len {
			netProto = ipv6.ProtocolNumber
		}

		tcpListener, err := w.CreateStackListener(addr, netProto)
		if err != nil {
			return fmt.Errorf("failed to create TCP listener on %v:%d: %w", listenerIP, listenerPort, err)
		}

		w.stackTCPListeners = append(w.stackTCPListeners, tcpListener)

		go w.acceptLoop(tcpListener, listenerTag)
	}
	return nil
}

func (w *WrappedStack) acceptLoop(listener *gonet.TCPListener, outboundTag string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go w.handleTCPConn(conn, outboundTag)
	}
}

func (w *WrappedStack) handleTCPConn(conn net.Conn, outboundTag string) {
	defer conn.Close()

	// Use the connection's local address as the destination,
	// representing the address the client was connecting to on the stack.
	tcpAddr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		return
	}

	dest := net.TCPDestination(net.IPAddress(tcpAddr.IP), net.Port(tcpAddr.Port))

	outboundConn, err := tagged.Dialer(w.ctx, dest, outboundTag)
	if err != nil {
		return
	}
	defer outboundConn.Close()

	// Bidirectional relay
	done := make(chan struct{})
	go func() {
		io.Copy(outboundConn, conn)
		close(done)
	}()
	io.Copy(conn, outboundConn)
	<-done
}

func (w *WrappedStack) CreateStackListener(addr tcpip.FullAddress, netProto tcpip.NetworkProtocolNumber) (*gonet.TCPListener, error) {
	return gonet.ListenTCP(w.stack, addr, netProto)
}
