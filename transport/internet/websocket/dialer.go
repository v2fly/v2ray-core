package websocket

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	gonet "net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/extension"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
)

// Dial dials a WebSocket connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialWebsocket(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial WebSocket").Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}

func dialWebsocket(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	wsSettings := streamSettings.ProtocolSettings.(*Config)

	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return internet.DialSystem(ctx, dest, streamSettings.SocketSettings)
		},
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		HandshakeTimeout: time.Second * 8,
	}

	protocol := "ws"

	securityEngine, err := security.CreateSecurityEngineFromSettings(ctx, streamSettings)
	if err != nil {
		return nil, newError("unable to create security engine").Base(err)
	}

	if securityEngine != nil {
		protocol = "wss"

		dialer.NetDialTLSContext = func(ctx context.Context, network, addr string) (gonet.Conn, error) {
			conn, err := dialer.NetDial(network, addr)
			if err != nil {
				return nil, newError("dial TLS connection failed").Base(err)
			}
			conn, err = securityEngine.Client(conn,
				security.OptionWithDestination{Dest: dest},
				security.OptionWithALPN{ALPNs: []string{"http/1.1"}})
			if err != nil {
				return nil, newError("unable to create security protocol client from security engine").Base(err)
			}
			return conn, nil
		}
	}

	host := dest.NetAddr()
	if (protocol == "ws" && dest.Port == 80) || (protocol == "wss" && dest.Port == 443) {
		host = dest.Address.String()
	}
	uri := protocol + "://" + host + wsSettings.GetNormalizedPath()

	if wsSettings.UseBrowserForwarding {
		var forwarder extension.BrowserForwarder
		err := core.RequireFeatures(ctx, func(Forwarder extension.BrowserForwarder) {
			forwarder = Forwarder
		})
		if err != nil {
			return nil, newError("cannot find browser forwarder service").Base(err)
		}
		if wsSettings.MaxEarlyData != 0 {
			return newRelayedConnectionWithDelayedDial(&dialerWithEarlyDataRelayed{
				forwarder: forwarder,
				uriBase:   uri,
				config:    wsSettings,
			}), nil
		}
		conn, err := forwarder.DialWebsocket(uri, nil)
		if err != nil {
			return nil, newError("cannot dial with browser forwarder service").Base(err)
		}
		return newRelayedConnection(conn), nil
	}

	if wsSettings.MaxEarlyData != 0 {
		return newConnectionWithDelayedDial(&dialerWithEarlyData{
			dialer:  dialer,
			uriBase: uri,
			config:  wsSettings,
		}), nil
	}

	conn, resp, err := dialer.Dial(uri, wsSettings.GetRequestHeader()) // nolint: bodyclose
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", uri, "): ", reason).Base(err)
	}

	return newConnection(conn, conn.RemoteAddr()), nil
}

type dialerWithEarlyData struct {
	dialer  *websocket.Dialer
	uriBase string
	config  *Config
}

func (d dialerWithEarlyData) Dial(earlyData []byte) (*websocket.Conn, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)

	earlydata := bytes.NewReader(earlyData)
	limitedEarlyDatareader := io.LimitReader(earlydata, int64(d.config.MaxEarlyData))
	n, encerr := io.Copy(base64EarlyDataEncoder, limitedEarlyDatareader)
	if encerr != nil {
		return nil, newError("websocket delayed dialer cannot encode early data").Base(encerr)
	}

	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}

	dialFunction := func() (*websocket.Conn, *http.Response, error) {
		return d.dialer.Dial(d.uriBase+earlyDataBuf.String(), d.config.GetRequestHeader())
	}

	if d.config.EarlyDataHeaderName != "" {
		dialFunction = func() (*websocket.Conn, *http.Response, error) {
			earlyDataStr := earlyDataBuf.String()
			currentHeader := d.config.GetRequestHeader()
			currentHeader.Set(d.config.EarlyDataHeaderName, earlyDataStr)
			return d.dialer.Dial(d.uriBase, currentHeader)
		}
	}

	conn, resp, err := dialFunction() // nolint: bodyclose
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", d.uriBase, ") with early data: ", reason).Base(err)
	}
	if n != int64(len(earlyData)) {
		if errWrite := conn.WriteMessage(websocket.BinaryMessage, earlyData[n:]); errWrite != nil {
			return nil, newError("failed to dial to (", d.uriBase, ") with early data as write of remainder early data failed: ").Base(err)
		}
	}
	return conn, nil
}

type dialerWithEarlyDataRelayed struct {
	forwarder extension.BrowserForwarder
	uriBase   string
	config    *Config
}

func (d dialerWithEarlyDataRelayed) Dial(earlyData []byte) (io.ReadWriteCloser, error) {
	earlyDataBuf := bytes.NewBuffer(nil)
	base64EarlyDataEncoder := base64.NewEncoder(base64.RawURLEncoding, earlyDataBuf)

	earlydata := bytes.NewReader(earlyData)
	limitedEarlyDatareader := io.LimitReader(earlydata, int64(d.config.MaxEarlyData))
	n, encerr := io.Copy(base64EarlyDataEncoder, limitedEarlyDatareader)
	if encerr != nil {
		return nil, newError("websocket delayed dialer cannot encode early data").Base(encerr)
	}

	if errc := base64EarlyDataEncoder.Close(); errc != nil {
		return nil, newError("websocket delayed dialer cannot encode early data tail").Base(errc)
	}

	dialFunction := func() (io.ReadWriteCloser, error) {
		return d.forwarder.DialWebsocket(d.uriBase+earlyDataBuf.String(), d.config.GetRequestHeader())
	}

	if d.config.EarlyDataHeaderName != "" {
		earlyDataStr := earlyDataBuf.String()
		currentHeader := d.config.GetRequestHeader()
		currentHeader.Set(d.config.EarlyDataHeaderName, earlyDataStr)
		return d.forwarder.DialWebsocket(d.uriBase, currentHeader)
	}

	conn, err := dialFunction()
	if err != nil {
		var reason string
		return nil, newError("failed to dial to (", d.uriBase, ") with early data: ", reason).Base(err)
	}
	if n != int64(len(earlyData)) {
		if _, errWrite := conn.Write(earlyData[n:]); errWrite != nil {
			return nil, newError("failed to dial to (", d.uriBase, ") with early data as write of remainder early data failed: ").Base(err)
		}
	}
	return conn, nil
}
