package websocket

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	http_proto "github.com/ghxhy/v2ray-core/v5/common/protocol/http"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
	v2tls "github.com/ghxhy/v2ray-core/v5/transport/internet/tls"
)

type requestHandler struct {
	path                string
	ln                  *Listener
	earlyDataEnabled    bool
	earlyDataHeaderName string
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:   4 * 1024,
	WriteBufferSize:  4 * 1024,
	HandshakeTimeout: time.Second * 4,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *requestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	responseHeader := http.Header{}

	var earlyData io.Reader
	if !h.earlyDataEnabled { // nolint: gocritic
		if request.URL.Path != h.path {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
	} else if h.earlyDataHeaderName != "" {
		if request.URL.Path != h.path {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		earlyDataStr := request.Header.Get(h.earlyDataHeaderName)
		earlyData = base64.NewDecoder(base64.RawURLEncoding, bytes.NewReader([]byte(earlyDataStr)))
		if strings.EqualFold("Sec-WebSocket-Protocol", h.earlyDataHeaderName) {
			responseHeader.Set(h.earlyDataHeaderName, earlyDataStr)
		}
	} else {
		if strings.HasPrefix(request.URL.RequestURI(), h.path) {
			earlyDataStr := request.URL.RequestURI()[len(h.path):]
			earlyData = base64.NewDecoder(base64.RawURLEncoding, bytes.NewReader([]byte(earlyDataStr)))
		} else {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
	}

	conn, err := upgrader.Upgrade(writer, request, responseHeader)
	if err != nil {
		newError("failed to convert to WebSocket connection").Base(err).WriteToLog()
		return
	}

	forwardedAddrs := http_proto.ParseXForwardedFor(request.Header)
	remoteAddr := conn.RemoteAddr()
	if len(forwardedAddrs) > 0 && forwardedAddrs[0].Family().IsIP() {
		remoteAddr = &net.TCPAddr{
			IP:   forwardedAddrs[0].IP(),
			Port: int(0),
		}
	}
	if earlyData == nil {
		h.ln.addConn(newConnection(conn, remoteAddr))
	} else {
		h.ln.addConn(newConnectionWithEarlyData(conn, remoteAddr, earlyData))
	}
}

type Listener struct {
	sync.Mutex
	server   http.Server
	listener net.Listener
	config   *Config
	addConn  internet.ConnHandler
}

func ListenWS(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	l := &Listener{
		addConn: addConn,
	}
	wsSettings := streamSettings.ProtocolSettings.(*Config)
	l.config = wsSettings
	if l.config != nil {
		if streamSettings.SocketSettings == nil {
			streamSettings.SocketSettings = &internet.SocketConfig{}
		}
		streamSettings.SocketSettings.AcceptProxyProtocol = l.config.AcceptProxyProtocol
	}
	var listener net.Listener
	var err error
	if port == net.Port(0) { // unix
		listener, err = internet.ListenSystem(ctx, &net.UnixAddr{
			Name: address.Domain(),
			Net:  "unix",
		}, streamSettings.SocketSettings)
		if err != nil {
			return nil, newError("failed to listen unix domain socket(for WS) on ", address).Base(err)
		}
		newError("listening unix domain socket(for WS) on ", address).WriteToLog(session.ExportIDToError(ctx))
	} else { // tcp
		listener, err = internet.ListenSystem(ctx, &net.TCPAddr{
			IP:   address.IP(),
			Port: int(port),
		}, streamSettings.SocketSettings)
		if err != nil {
			return nil, newError("failed to listen TCP(for WS) on ", address, ":", port).Base(err)
		}
		newError("listening TCP(for WS) on ", address, ":", port).WriteToLog(session.ExportIDToError(ctx))
	}

	if streamSettings.SocketSettings != nil && streamSettings.SocketSettings.AcceptProxyProtocol {
		newError("accepting PROXY protocol").AtWarning().WriteToLog(session.ExportIDToError(ctx))
	}

	if config := v2tls.ConfigFromStreamSettings(streamSettings); config != nil {
		if tlsConfig := config.GetTLSConfig(); tlsConfig != nil {
			listener = tls.NewListener(listener, tlsConfig)
		}
	}

	l.listener = listener
	useEarlyData := false
	earlyDataHeaderName := ""
	if wsSettings.MaxEarlyData != 0 {
		useEarlyData = true
		earlyDataHeaderName = wsSettings.EarlyDataHeaderName
	}

	l.server = http.Server{
		Handler: &requestHandler{
			path:                wsSettings.GetNormalizedPath(),
			ln:                  l,
			earlyDataEnabled:    useEarlyData,
			earlyDataHeaderName: earlyDataHeaderName,
		},
		ReadHeaderTimeout: time.Second * 4,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	go func() {
		if err := l.server.Serve(l.listener); err != nil {
			newError("failed to serve http for WebSocket").Base(err).AtWarning().WriteToLog(session.ExportIDToError(ctx))
		}
	}()

	return l, err
}

// Addr implements net.Listener.Addr().
func (ln *Listener) Addr() net.Addr {
	return ln.listener.Addr()
}

// Close implements net.Listener.Close().
func (ln *Listener) Close() error {
	return ln.listener.Close()
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, ListenWS))
}
