package udp

import (
	"errors"
	"fmt"
	gonet "net"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type Server struct {
	Port         net.Port
	MsgProcessor func(msg []byte) []byte

	access sync.Mutex
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
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	server.Port = net.Port(localAddr.Port)
	fmt.Println("UDP server started on port ", server.Port)

	server.access.Lock()
	server.conn = conn
	server.access.Unlock()

	go server.handleConnection(conn)

	return net.UDPDestination(net.IPAddress(localAddr.IP), net.Port(localAddr.Port)), nil
}

func (server *Server) handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 2*1024)
	for {
		nBytes, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// A closed connection never recovers, so continuing the loop would
			// spin on the same error forever instead of ending the goroutine.
			if errors.Is(err, gonet.ErrClosed) {
				return
			}
			fmt.Printf("Failed to read from UDP: %v\n", err)
			continue
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

	if server.conn == nil {
		return nil
	}
	conn := server.conn
	server.conn = nil
	return conn.Close()
}
