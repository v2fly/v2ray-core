package udp

import (
	"fmt"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type Server struct {
	Port         net.Port
	MsgProcessor func(msg []byte) []byte

	access sync.Mutex
	closed bool
	conn   *net.UDPConn
}

func (server *Server) Start() (net.Destination, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(server.Port),
		Zone: "",
	})
	if err != nil {
		return net.Destination{}, err
	}
	server.Port = net.Port(conn.LocalAddr().(*net.UDPAddr).Port)
	fmt.Println("UDP server started on port ", server.Port)

	server.access.Lock()
	server.closed = false
	server.conn = conn
	server.access.Unlock()

	go server.handleConnection(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return net.UDPDestination(net.IPAddress(localAddr.IP), net.Port(localAddr.Port)), nil
}

func (server *Server) isClosed() bool {
	server.access.Lock()
	defer server.access.Unlock()
	return server.closed
}

func (server *Server) handleConnection(conn *net.UDPConn) {
	for {
		buffer := make([]byte, 2*1024)
		nBytes, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if server.isClosed() {
				return
			}
			fmt.Printf("Failed to read from UDP: %v\n", err)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		response := server.MsgProcessor(buffer[:nBytes])
		if _, err := conn.WriteToUDP(response, addr); err != nil {
			fmt.Println("Failed to write to UDP: ", err.Error())
		}
	}
}

func (server *Server) Close() error {
	server.access.Lock()
	defer server.access.Unlock()

	server.closed = true
	if server.conn == nil {
		return nil
	}
	return server.conn.Close()
}
