package udp

import (
	"fmt"
	"golang.org/x/net/context"

	"v2ray.com/core/common/net"
)

type Server struct {
	Port         net.Port
	MsgProcessor func(msg []byte) []byte
	conn         *net.UDPConn
	cancel       context.CancelFunc
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

	ctx, cancel := context.WithCancel(context.Background())
	server.conn = conn
	server.cancel = cancel

	go server.handleConnection(ctx, conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return net.UDPDestination(net.IPAddress(localAddr.IP), net.Port(localAddr.Port)), nil
}

func (server *Server) handleConnection(ctx context.Context, conn *net.UDPConn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buffer := make([]byte, 2*1024)
			nBytes, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("Failed to read from UDP: %v\n", err)
				continue
			}

			response := server.MsgProcessor(buffer[:nBytes])
			if _, err := conn.WriteToUDP(response, addr); err != nil {
				fmt.Println("Failed to write to UDP: ", err.Error())
			}
		}
	}
}

func (server *Server) Close() error {
	server.cancel()
	return server.conn.Close()
}
