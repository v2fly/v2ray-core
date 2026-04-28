package hysteria2

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/apernet/quic-go"
	"github.com/apernet/quic-go/http3"
	"github.com/apernet/quic-go/quicvarint"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2/congestion"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2/congestion/bbr"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2/salamander"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

type h3sHandler struct {
	sync.Mutex

	config  *Config
	addConn internet.ConnHandler
	conn    *quic.Conn

	authenticated bool
	udpSM         *udpSessionManager
}

func (h *h3sHandler) AuthHTTP(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	if h.authenticated {
		w.Header().Set(ResponseHeaderUDPEnabled, strconv.FormatBool(true))
		w.Header().Set(CommonHeaderCCRX, strconv.FormatUint(h.config.BrutalRxMbps*1000*1000/8, 10))
		w.Header().Set(CommonHeaderPadding, AuthResponsePadding.String())
		w.WriteHeader(StatusAuthOK)
		return
	}

	auth := r.Header.Get(RequestHeaderAuth)
	rx, _ := strconv.ParseUint(r.Header.Get(CommonHeaderCCRX), 10, 64)

	if auth == h.config.Auth {
		h.authenticated = true

		h.udpSM = &udpSessionManager{
			conn: h.conn,
			m:    make(map[uint32]*InterConn),

			addConn: h.addConn,
		}
		go h.udpSM.clean()
		go h.udpSM.run()

		switch h.config.Congestion {
		case "reno":
		case "bbr":
			congestion.UseBBR(h.conn, bbr.Profile(h.config.BbrProfile))
		case "", "brutal":
			if h.config.BrutalTxMbps > 0 && rx > 0 {
				congestion.UseBrutal(h.conn, min(h.config.BrutalTxMbps*1000*1000/8, rx))
			} else {
				congestion.UseBBR(h.conn, bbr.Profile(h.config.BbrProfile))
			}
		case "force-brutal":
			congestion.UseBrutal(h.conn, h.config.BrutalTxMbps*1000*1000/8)
		default:
			panic(h.config.Congestion)
		}

		w.Header().Set(ResponseHeaderUDPEnabled, strconv.FormatBool(true))
		w.Header().Set(CommonHeaderCCRX, strconv.FormatUint(h.config.BrutalRxMbps*1000*1000/8, 10))
		w.Header().Set(CommonHeaderPadding, AuthResponsePadding.String())
		w.WriteHeader(StatusAuthOK)
	}
}

func (h *h3sHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.Host == URLHost && r.URL.Path == URLPath {
		h.AuthHTTP(w, r)
		if h.authenticated {
			return
		}
	}
	http.NotFound(w, r)
}

func (h *h3sHandler) StreamDispatcher(ft http3.FrameType, stream *quic.Stream, err error) (bool, error) {
	if err != nil || !h.authenticated {
		return false, nil
	}

	switch ft {
	case FrameTypeTCPRequest:
		if _, err := quicvarint.Read(quicvarint.NewReader(stream)); err != nil {
			return false, err
		}

		h.addConn(&interConn{
			stream: stream,
			local:  h.conn.LocalAddr(),
			remote: h.conn.RemoteAddr(),
		})
		return true, nil
	default:
		return false, nil
	}
}

type Listener struct {
	config   *Config
	addConn  internet.ConnHandler
	pktConn  net.PacketConn
	tr       *quic.Transport
	listener *quic.Listener
}

func (l *Listener) handleClient(conn *quic.Conn) {
	handler := &h3sHandler{
		config:  l.config,
		addConn: l.addConn,
		conn:    conn,
	}
	h3s := http3.Server{
		Handler:          handler,
		StreamDispatcher: handler.StreamDispatcher,
	}
	_ = h3s.ServeQUICConn(conn)
	_ = conn.CloseWithError(closeErrCodeOK, "")
}

func (l *Listener) keepAccepting() {
	for {
		conn, err := l.listener.Accept(context.Background())
		if err != nil {
			break
		}
		go l.handleClient(conn)
	}
}

func (l *Listener) Addr() net.Addr {
	return l.pktConn.LocalAddr()
}

func (l *Listener) Close() error {
	return errors.Join(l.listener.Close(), l.tr.Close(), l.pktConn.Close())
}

func Listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	if tls.ConfigFromStreamSettings(streamSettings) == nil {
		return nil, newError("tls is nil")
	}

	if address.Family().IsDomain() {
		return nil, newError("address is domain")
	}

	config := streamSettings.ProtocolSettings.(*Config)

	pktConn, err := internet.ListenSystemPacket(context.Background(), &net.UDPAddr{IP: address.IP(), Port: int(port)}, streamSettings.SocketSettings)
	if err != nil {
		return nil, err
	}
	if config.Salamander != nil {
		obfs, err := salamander.NewSalamanderObfuscator([]byte(*config.Salamander))
		if err != nil {
			return nil, err
		}
		pktConn = salamander.WrapPacketConn(pktConn, obfs)
	}

	tlsConfig := tls.ConfigFromStreamSettings(streamSettings).GetTLSConfig()
	quicConfig := &quic.Config{
		InitialStreamReceiveWindow:     config.InitialStreamReceiveWindow,
		MaxStreamReceiveWindow:         config.MaxStreamReceiveWindow,
		InitialConnectionReceiveWindow: config.InitialConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     config.MaxConnectionReceiveWindow,
		MaxIdleTimeout:                 time.Duration(config.MaxIdleTimeout) * time.Second,
		MaxIncomingStreams:             config.MaxIncomingStreams,
		DisablePathMTUDiscovery:        config.DisablePathMTUDiscovery || (runtime.GOOS != "linux" && runtime.GOOS != "windows" && runtime.GOOS != "darwin"),
		EnableDatagrams:                true,
		MaxDatagramFrameSize:           MaxDatagramFrameSize,
		AssumePeerMaxDatagramFrameSize: MaxDatagramFrameSize,
		DisablePathManager:             true,
	}
	tr := &quic.Transport{Conn: pktConn}
	listener, err := tr.Listen(tlsConfig, quicConfig)
	if err != nil {
		_ = tr.Close()
		_ = pktConn.Close()
		return nil, err
	}

	l := &Listener{
		config:   config,
		addConn:  handler,
		pktConn:  pktConn,
		tr:       tr,
		listener: listener,
	}

	go l.keepAccepting()

	return l, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
