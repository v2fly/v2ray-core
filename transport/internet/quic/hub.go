package quic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	core "github.com/v2fly/v2ray-core/v5"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

// Listener is an internet.Listener that listens for TCP connections.
type Listener struct {
	rawConn  *sysConn
	listener *quic.Listener
	done     *done.Instance
	addConn  internet.ConnHandler
}

func (l *Listener) acceptStreams(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			newError("failed to accept stream").Base(err).WriteToLog()
			select {
			case <-conn.Context().Done():
				return
			case <-l.done.Wait():
				if err := conn.CloseWithError(0, ""); err != nil {
					newError("failed to close connection").Base(err).WriteToLog()
				}
				return
			default:
				time.Sleep(time.Second)
				continue
			}
		}

		conn := &interConn{
			stream: stream,
			local:  conn.LocalAddr(),
			remote: conn.RemoteAddr(),
		}

		l.addConn(conn)
	}
}

func (l *Listener) keepAccepting() {
	for {
		conn, err := l.listener.Accept(context.Background())
		if err != nil {
			newError("failed to accept QUIC connections").Base(err).WriteToLog()
			if l.done.Done() {
				break
			}
			time.Sleep(time.Second)
			continue
		}
		go l.acceptStreams(conn)
	}
}

// Addr implements internet.Listener.Addr.
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close implements internet.Listener.Close.
func (l *Listener) Close() error {
	l.done.Close()
	l.listener.Close()
	l.rawConn.Close()
	return nil
}

// Listen creates a new Listener based on configurations.
func Listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	if address.Family().IsDomain() {
		return nil, newError("domain address is not allows for listening quic")
	}

	tlsConfig := tls.ConfigFromStreamSettings(streamSettings)
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			Certificate: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.DNSNames(internalDomain), cert.CommonName(internalDomain)))},
		}
	}

	config := streamSettings.ProtocolSettings.(*Config)
	rawConn, err := internet.ListenSystemPacket(context.Background(), &net.UDPAddr{
		IP:   address.IP(),
		Port: int(port),
	}, streamSettings.SocketSettings)
	if err != nil {
		return nil, err
	}

	quicConfig := &quic.Config{
		HandshakeIdleTimeout:  time.Second * 8,
		MaxIdleTimeout:        time.Second * 45,
		MaxIncomingStreams:    32,
		MaxIncomingUniStreams: -1,
		KeepAlivePeriod:       time.Second * 15,
	}

	conn, err := wrapSysConn(rawConn.(*net.UDPConn), config)
	if err != nil {
		conn.Close()
		return nil, err
	}

	connectionIDLength := 4

	if config.ConnectionIdLength != nil {
		connectionIDLength = int(*config.ConnectionIdLength)
	}

	tr := quic.Transport{
		Conn:               conn,
		ConnectionIDLength: connectionIDLength,
	}

	defaultNextProtos := []string{"h3"}

	{
		// Until v5.61.0, server will have a default alpn of h3,h2,http/1.1 then h3 only
		extractXY := func(s string) (x, y string, err error) {
			parts := strings.SplitN(s, ".", 3)
			if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
				return "", "", fmt.Errorf("invalid format: %q", s)
			}
			return parts[0], parts[1], nil
		}
		func() {
			if major, middle, err := extractXY(core.Version()); err != nil {
				return
			} else {
				if major != "5" {
					return
				}
				if middleInt, err := strconv.ParseInt(middle, 10, 64); err != nil || middleInt > 60 {
					return
				}
			}
			defaultNextProtos = []string{"h3", "h2", "http/1.1"}
		}()
	}

	qListener, err := tr.Listen(tlsConfig.GetTLSConfig(tls.WithNextProto(defaultNextProtos...)), quicConfig)
	if err != nil {
		conn.Close()
		return nil, err
	}

	listener := &Listener{
		done:     done.New(),
		rawConn:  conn,
		listener: qListener,
		addConn:  handler,
	}

	go listener.keepAccepting()

	return listener, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
