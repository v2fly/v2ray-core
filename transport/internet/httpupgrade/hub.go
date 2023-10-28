package httpupgrade

import (
	"bufio"
	"context"
	"net/http"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"
)

type server struct {
	addConn        internet.ConnHandler
	innnerListener net.Listener
}

func (s *server) Close() error {
	return s.innnerListener.Close()
}

func (s *server) Addr() net.Addr {
	return nil
}

func (s *server) Handle(conn net.Conn) (internet.Connection, error) {
	connReader := bufio.NewReader(conn)
	req, err := http.ReadRequest(connReader)
	if err != nil {
		return nil, err
	}
	connection := strings.ToLower(req.Header.Get("Connection"))
	upgrade := strings.ToLower(req.Header.Get("Upgrade"))
	if connection != "upgrade" || upgrade != "websocket" {
		_ = conn.Close()
		return nil, newError("unrecognized request")
	}
	resp := &http.Response{
		Status:     "101 Switching Protocols",
		StatusCode: 101,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	resp.Header.Set("Connection", "upgrade")
	resp.Header.Set("Upgrade", "websocket")
	err = resp.Write(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return internet.Connection(conn), nil
}

func (s *server) keepAccepting() {
	for {
		conn, err := s.innnerListener.Accept()
		if err != nil {
			return
		}
		handledConn, err := s.Handle(conn)
		if err != nil {
			newError("failed to handle request").Base(err).WriteToLog()
			continue
		}
		s.addConn(handledConn)
	}
}

func listenHTTPUpgrade(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)
	_ = transportConfiguration
	serverInstance := &server{addConn: addConn}

	listener, err := transportcommon.ListenWithSecuritySettings(ctx, address, port, streamSettings)
	if err != nil {
		return nil, newError("failed to listen on ", address, ":", port).Base(err)
	}
	serverInstance.innnerListener = listener
	go serverInstance.keepAccepting()
	return serverInstance, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, listenHTTPUpgrade))
}
