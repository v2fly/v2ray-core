package hysteria2

import (
	"context"
	gotls "crypto/tls"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/apernet/quic-go"
	"github.com/apernet/quic-go/http3"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2/congestion"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2/congestion/bbr"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

type client struct {
	sync.Mutex

	dest         net.Destination
	config       *Config
	tlsConfig    *gotls.Config
	socketConfig *internet.SocketConfig

	conn    *quic.Conn
	tr      *quic.Transport
	pktConn net.PacketConn
	udpSM   *udpSessionManager
}

func (c *client) status() status {
	if c.conn == nil {
		return StatusNull
	}
	select {
	case <-c.conn.Context().Done():
		return StatusInactive
	default:
		return StatusActive
	}
}

func (c *client) close() {
	c.conn.CloseWithError(closeErrCodeOK, "")
	c.tr.Close()
	c.pktConn.Close()
	c.conn = nil
	c.tr = nil
	c.pktConn = nil
	c.udpSM = nil
}

func (c *client) dial(ctx context.Context) error {
	status := c.status()
	if status == StatusActive {
		return nil
	}
	if status == StatusInactive {
		c.close()
	}

	udpAddr, err := net.ResolveUDPAddr("udp", c.dest.NetAddr())
	if err != nil {
		return err
	}
	pktConn, err := internet.ListenSystemPacket(ctx, &net.UDPAddr{Port: 0}, c.socketConfig)
	if err != nil {
		return err
	}

	tr := &quic.Transport{Conn: pktConn}

	var conn *quic.Conn
	rt := &http3.Transport{
		TLSClientConfig: c.tlsConfig,
		QUICConfig: &quic.Config{
			InitialStreamReceiveWindow:     c.config.InitialStreamReceiveWindow,
			MaxStreamReceiveWindow:         c.config.MaxStreamReceiveWindow,
			InitialConnectionReceiveWindow: c.config.InitialConnectionReceiveWindow,
			MaxConnectionReceiveWindow:     c.config.MaxConnectionReceiveWindow,
			MaxIdleTimeout:                 time.Duration(c.config.MaxIdleTimeout) * time.Second,
			KeepAlivePeriod:                time.Duration(c.config.KeepAlivePeriod) * time.Second,
			DisablePathMTUDiscovery:        c.config.DisablePathMTUDiscovery || (runtime.GOOS != "linux" && runtime.GOOS != "windows" && runtime.GOOS != "darwin"),
			EnableDatagrams:                true,
			MaxDatagramFrameSize:           MaxDatagramFrameSize,
			OmitMaxDatagramFrameSize:       true,
			DisablePathManager:             true,
		},
		Dial: func(ctx context.Context, _ string, tlsCfg *gotls.Config, cfg *quic.Config) (*quic.Conn, error) {
			qc, err := tr.DialEarly(ctx, udpAddr, tlsCfg, cfg)
			if err != nil {
				return nil, err
			}
			conn = qc
			return qc, nil
		},
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "https",
			Host:   URLHost,
			Path:   URLPath,
		},
		Header: http.Header{
			RequestHeaderAuth:   []string{c.config.Auth},
			CommonHeaderCCRX:    []string{strconv.FormatUint(c.config.BrutalRxMbps*1000*1000/8, 10)},
			CommonHeaderPadding: []string{AuthRequestPadding.String()},
		},
	}
	resp, err := rt.RoundTrip(req)
	if err != nil {
		if conn != nil {
			_ = conn.CloseWithError(closeErrCodeProtocolError, "")
		}
		_ = tr.Close()
		_ = pktConn.Close()
		return err
	}
	if resp.StatusCode != StatusAuthOK {
		_ = conn.CloseWithError(closeErrCodeProtocolError, "")
		_ = tr.Close()
		_ = pktConn.Close()
		return newError("auth failed code ", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// udp, _ := strconv.ParseBool(resp.Header.Get(ResponseHeaderUDPEnabled))
	rx, _ := strconv.ParseUint(resp.Header.Get(CommonHeaderCCRX), 10, 64)

	switch c.config.Congestion {
	case "reno":
	case "bbr":
		congestion.UseBBR(conn, bbr.Profile(c.config.BbrProfile))
	case "", "brutal":
		if c.config.BrutalTxMbps == 0 || rx == 0 {
			congestion.UseBBR(conn, bbr.Profile(c.config.BbrProfile))
		} else {
			congestion.UseBrutal(conn, min(c.config.BrutalTxMbps*1000*1000/8, rx))
		}
	case "force-brutal":
		congestion.UseBrutal(conn, c.config.BrutalTxMbps*1000*1000/8)
	default:
		panic(c.config.Congestion)
	}

	c.pktConn = pktConn
	c.tr = tr
	c.conn = conn
	c.udpSM = &udpSessionManager{
		conn: conn,
		m:    make(map[uint32]*InterConn),
		next: 1,
	}
	go c.udpSM.run()

	return nil
}

func (c *client) tcp(ctx context.Context) (net.Conn, error) {
	c.Lock()
	defer c.Unlock()

	err := c.dial(ctx)
	if err != nil {
		return nil, err
	}

	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	return &interConn{
		stream: stream,
		local:  c.conn.LocalAddr(),
		remote: c.conn.RemoteAddr(),

		client: true,
	}, nil
}

func (c *client) udp(ctx context.Context) (net.Conn, error) {
	c.Lock()
	defer c.Unlock()

	err := c.dial(ctx)
	if err != nil {
		return nil, err
	}

	return c.udpSM.udp()
}

func (c *client) clean() {
	c.Lock()
	if c.status() == StatusInactive {
		c.close()
	}
	c.Unlock()
}

type dialerConf struct {
	net.Destination
	*internet.MemoryStreamConfig
}

type clientManager struct {
	sync.RWMutex
	m map[dialerConf]*client
}

func (m *clientManager) clean() {
	ticker := time.NewTicker(idleCleanupInterval)
	for range ticker.C {
		m.RLock()
		for _, c := range m.m {
			c.clean()
		}
		m.RUnlock()
	}
}

var manager *clientManager

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	tlsConfig := tls.ConfigFromStreamSettings(streamSettings)
	if tlsConfig == nil {
		return nil, newError("tls config is nil")
	}

	datagram := DatagramFromContext(ctx)

	manager.RLock()
	c := manager.m[dialerConf{dest, streamSettings}]
	manager.RUnlock()

	if c == nil {
		manager.Lock()
		c = manager.m[dialerConf{dest, streamSettings}]
		if c == nil {
			c = &client{
				dest:         dest,
				config:       streamSettings.ProtocolSettings.(*Config),
				tlsConfig:    tlsConfig.GetTLSConfig(),
				socketConfig: streamSettings.SocketSettings,
			}
			manager.m[dialerConf{dest, streamSettings}] = c
		}
		manager.Unlock()
	}

	if datagram {
		return c.udp(ctx)
	}
	return c.tcp(ctx)
}

func init() {
	manager = &clientManager{
		m: make(map[dialerConf]*client),
	}
	go manager.clean()
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
