package httpupgrade

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"
)

type server struct {
	config *Config

	addConn        internet.ConnHandler
	innnerListener net.Listener
}

func (s *server) Close() error {
	return s.innnerListener.Close()
}

func (s *server) Addr() net.Addr {
	return nil
}

func (s *server) Handle(conn net.Conn) {
	upgradedConn, err := s.upgrade(conn)
	if err != nil {
		conn.Close()
		newError("failed to handle request").Base(err).WriteToLog()
		return
	}
	s.addConn(upgradedConn)
}

// upgrade execute a fake websocket upgrade process and return the available connection
func (s *server) upgrade(conn net.Conn) (internet.Connection, error) {
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
	if s.config.MaxEarlyData != 0 {
		if s.config.EarlyDataHeaderName == "" {
			return nil, newError("EarlyDataHeaderName is not set")
		}
		earlyData := req.Header.Get(s.config.EarlyDataHeaderName)
		if earlyData != "" {
			earlyDataBytes, err := base64.URLEncoding.DecodeString(earlyData)
			if err != nil {
				return nil, err
			}
			return newConnectionWithPendingRead(conn, conn.RemoteAddr(), bytes.NewReader(earlyDataBytes)), nil
		}
	}
	return internet.Connection(conn), nil
}

func (s *server) keepAccepting() {
	for {
		conn, err := s.innnerListener.Accept()
		if err != nil {
			return
		}
		go s.Handle(conn)
	}
}

func listenHTTPUpgrade(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)
	serverInstance := &server{config: transportConfiguration, addConn: addConn}

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
