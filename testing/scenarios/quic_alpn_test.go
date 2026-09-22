package scenarios

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"
	"text/template"
	"time"

	quicgo "github.com/quic-go/quic-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
	"github.com/v2fly/v2ray-core/v5/features/extension"
	"github.com/v2fly/v2ray-core/v5/testing/servers/tcp"
	"github.com/v2fly/v2ray-core/v5/testing/servers/udp"
	v2tls "github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

func TestQUICALPNConfigProxy(t *testing.T) {
	tests := []struct {
		name           string
		server, client []string
	}{
		{name: "matching", server: []string{"h3"}, client: []string{"h3"}},
		{name: "legacy_client", server: []string{"h3"}, client: []string{"http/1.1"}},
		{name: "multiple_mismatches", server: []string{"h3", "h2"}, client: []string{"legacy", "other"}},
		{name: "later_overlap", server: []string{"h3", "h2"}, client: []string{"legacy", "h2"}},
		{name: "default_server_alpn", client: []string{"legacy"}},
		{name: "default_client_alpn", server: []string{"legacy"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := startQUICALPNServer(t, tt.server)
			echo := &tcp.Server{MsgProcessor: xor}
			dest, err := echo.Start()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { echo.Close() })

			socksPort := tcp.PickPort()
			clientTLS := &v2tls.Config{
				ServerName: "quic-alpn.test", NextProtocol: tt.client,
				DisableSystemRoot: true, EnableSessionResumption: true,
				Certificate: []*v2tls.Certificate{{Certificate: server.certPEM, Usage: v2tls.Certificate_AUTHORITY_VERIFY}},
			}
			startQUICALPNInstance(t, server.manager, "quic_client", "config/quic_alpn_client.json.tmpl", server.port, socksPort, clientTLS)

			// Exercise the configured SOCKS inbound, VMess outbound, QUIC listener,
			// VMess inbound and freedom outbound, including a verified TLS certificate.
			if err := testTCPConnViaSocks(socksPort, dest.Port, 4096, 10*time.Second)(); err != nil {
				t.Fatalf("initial proxy connection: %v", err)
			}
			var transfers errgroup.Group
			for range 4 {
				transfers.Go(testTCPConnViaSocks(socksPort, dest.Port, 64*1024, 10*time.Second))
			}
			if err := transfers.Wait(); err != nil {
				t.Fatalf("concurrent proxy streams: %v", err)
			}
		})
	}
}

func TestQUICALPNConfigNegotiation(t *testing.T) {
	server := startQUICALPNServer(t, []string{"h3", "h2"})
	// Each dial creates a fresh QUIC connection. The V2Ray outbound pool is
	// deliberately bypassed so every ClientHello reaches the same live server.
	tests := []struct {
		name   string
		protos []string
		want   string
	}{
		{name: "first_mismatch", protos: []string{"legacy-a", "other"}, want: "legacy-a"},
		{name: "server_preference_after_mismatch", protos: []string{"h2", "h3"}, want: "h3"},
		{name: "second_mismatch", protos: []string{"legacy-b", "legacy-a"}, want: "legacy-b"},
		{name: "later_overlap_after_mismatch", protos: []string{"legacy-a", "h2"}, want: "h2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.clientTLS.Clone()
			config.NextProtos = tt.protos
			conn := dialQUICALPNServer(t, server.port, config)
			if got := conn.ConnectionState().TLS.NegotiatedProtocol; got != tt.want {
				t.Errorf("negotiated ALPN = %q, want %q", got, tt.want)
			}
		})
	}
	t.Run("concurrent_clients", func(t *testing.T) {
		for i := range 8 {
			t.Run(fmt.Sprintf("client_%d", i), func(t *testing.T) {
				t.Parallel()
				want := fmt.Sprintf("legacy-%d", i)
				config := server.clientTLS.Clone()
				config.NextProtos = []string{want}
				conn := dialQUICALPNServer(t, server.port, config)
				if got := conn.ConnectionState().TLS.NegotiatedProtocol; got != want {
					t.Errorf("negotiated ALPN = %q, want %q", got, want)
				}
			})
		}
	})
	// Confirm the shared server preference survives all concurrent fallbacks.
	config := server.clientTLS.Clone()
	config.NextProtos = []string{"h2", "h3"}
	conn := dialQUICALPNServer(t, server.port, config)
	if got := conn.ConnectionState().TLS.NegotiatedProtocol; got != "h3" {
		t.Errorf("server preference after concurrent connections = %q, want h3", got)
	}
}

func TestQUICALPNConfigSessionResumption(t *testing.T) {
	server := startQUICALPNServer(t, []string{"h3"})
	cache := &quicALPNSessionCache{
		ClientSessionCache: tls.NewLRUClientSessionCache(1),
		stored:             make(chan struct{}, 1),
	}
	config := server.clientTLS.Clone()
	config.NextProtos = []string{"legacy"}
	config.ClientSessionCache = cache
	for i := range 3 {
		t.Run(fmt.Sprintf("connection_%d", i), func(t *testing.T) {
			conn := dialQUICALPNServer(t, server.port, config)
			state := conn.ConnectionState().TLS
			if state.DidResume != (i > 0) {
				t.Errorf("DidResume = %v, want %v", state.DidResume, i > 0)
			}
			if state.NegotiatedProtocol != "legacy" {
				t.Errorf("negotiated ALPN = %q, want legacy", state.NegotiatedProtocol)
			}
			select {
			case <-cache.stored:
			case <-time.After(5 * time.Second):
				t.Fatal("configured server did not issue a session ticket")
			}
		})
	}
}

type quicALPNRuntime struct {
	manager   extension.InstanceManagement
	port      v2net.Port
	certPEM   []byte
	clientTLS *tls.Config
}

func startQUICALPNServer(t *testing.T, protos []string) *quicALPNRuntime {
	t.Helper()
	certificate, err := cert.Generate(nil, cert.DNSNames("quic-alpn.test"))
	if err != nil {
		t.Fatal(err)
	}
	certPEM, keyPEM := certificate.ToPEM()
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certPEM) {
		t.Fatal("could not trust the test certificate")
	}
	coreInst, manager := NewInstanceManagerCoreInstance()
	t.Cleanup(func() {
		if err := coreInst.Close(); err != nil {
			t.Errorf("close instance manager: %v", err)
		}
	})
	serverPort := udp.PickPort()
	serverTLS := &v2tls.Config{
		NextProtocol: protos, EnableSessionResumption: true,
		Certificate: []*v2tls.Certificate{{Certificate: certPEM, Key: keyPEM}},
	}
	startQUICALPNInstance(t, manager, "quic_server", "config/quic_alpn_server.json.tmpl", serverPort, 0, serverTLS)
	return &quicALPNRuntime{
		manager: manager, port: serverPort, certPEM: certPEM,
		clientTLS: &tls.Config{ServerName: "quic-alpn.test", RootCAs: roots, MinVersion: tls.VersionTLS13},
	}
}

func startQUICALPNInstance(t *testing.T, manager extension.InstanceManagement, name, configPath string, serverPort, clientPort v2net.Port, tlsConfig *v2tls.Config) {
	t.Helper()
	// Generate certificates and choose ports for each test, keeping the complete
	// client/server configurations in fixtures without fixed ports or expiring keys.
	tlsJSON, err := protojson.Marshal(tlsConfig)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := template.ParseFiles(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var config bytes.Buffer
	if err := fixture.Execute(&config, struct {
		ServerPort, ClientPort uint16
		TLSSettings            string
	}{uint16(serverPort), uint16(clientPort), string(tlsJSON)}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := manager.AddInstance(ctx, name, config.Bytes(), "jsonv5"); err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	t.Cleanup(func() {
		if err := manager.StopInstance(ctx, name); err != nil {
			t.Errorf("stop %s: %v", name, err)
		}
		if err := manager.UntrackInstance(ctx, name); err != nil {
			t.Errorf("untrack %s: %v", name, err)
		}
	})
	if err := manager.StartInstance(ctx, name); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
}

func dialQUICALPNServer(t *testing.T, port v2net.Port, config *tls.Config) *quicgo.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := quicgo.DialAddr(ctx, "127.0.0.1:"+port.String(), config, &quicgo.Config{HandshakeIdleTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("QUIC handshake with configured server: %v", err)
	}
	t.Cleanup(func() { conn.CloseWithError(0, "test complete") })
	return conn
}

type quicALPNSessionCache struct {
	tls.ClientSessionCache
	stored chan struct{}
}

func (c *quicALPNSessionCache) Put(key string, state *tls.ClientSessionState) {
	c.ClientSessionCache.Put(key, state)
	if state != nil {
		select {
		case c.stored <- struct{}{}:
		default:
		}
	}
}
