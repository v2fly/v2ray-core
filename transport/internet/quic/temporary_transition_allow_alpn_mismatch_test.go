package quic

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
)

func TestExtractXY(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		x, y    string
		wantErr bool
	}{
		{name: "release", input: "5.54.1", x: "5", y: "54"},
		{name: "without_patch", input: "5.54", x: "5", y: "54"},
		{name: "empty_patch", input: "5.54.", x: "5", y: "54"},
		{name: "prerelease_patch", input: "5.54.1-rc.2+build.3", x: "5", y: "54"},
		{name: "extra_components", input: "5.54.1.2.3", x: "5", y: "54"},
		{name: "empty_trailing_components", input: "5.54...", x: "5", y: "54"},
		{name: "zero", input: "0.0.0", x: "0", y: "0"},
		{name: "leading_zeroes", input: "05.054.1", x: "05", y: "054"},
		// Extraction is structural; numeric validation belongs to the version gate.
		{name: "nonnumeric", input: "major.middle.patch", x: "major", y: "middle"},
		{name: "signed", input: "+5.-54.1", x: "+5", y: "-54"},
		{name: "whitespace_preserved", input: " 5.54 .1", x: " 5", y: "54 "},
		{name: "unicode", input: "五.五十四.一", x: "五", y: "五十四"},
		{name: "empty", input: "", wantErr: true},
		{name: "major_only", input: "5", wantErr: true},
		{name: "no_dot", input: "5541", wantErr: true},
		{name: "dot_only", input: ".", wantErr: true},
		{name: "dots_only", input: "...", wantErr: true},
		{name: "missing_major", input: ".54.1", wantErr: true},
		{name: "missing_middle", input: "5..1", wantErr: true},
		{name: "trailing_dot", input: "5.", wantErr: true},
		{name: "missing_first_two", input: "..1", wantErr: true},
		{name: "quoted_invalid_input", input: "\"\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, err := extractXY(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractXY(%q) error = %v, want error %v", tt.input, err, tt.wantErr)
			}
			if x != tt.x || y != tt.y {
				t.Errorf("extractXY(%q) = (%q, %q), want (%q, %q)", tt.input, x, y, tt.x, tt.y)
			}
			if tt.wantErr && err.Error() != fmt.Sprintf("invalid format: %q", tt.input) {
				t.Errorf("error does not identify the invalid input: %v", err)
			}
		})
	}
}

// The repository version is inside the transition window. To exercise the gate
// for other versions without changing production code or package globals, run:
//
//	go test ./transport/internet/quic -run '^TestSetupTLSConfigForALPNMismatchVersionGate$' \
//	  -ldflags '-X github.com/v2fly/v2ray-core/v5.version=5.61.0 -X github.com/v2fly/v2ray-core/v5/transport/internet/quic.alpnMismatchExpected=disabled'
//
// Use "enabled" for versions that should still receive the workaround.
var alpnMismatchExpected = "enabled"

func TestSetupTLSConfigForALPNMismatchVersionGate(t *testing.T) {
	if alpnMismatchExpected != "enabled" && alpnMismatchExpected != "disabled" {
		t.Fatalf("invalid ALPN workaround expectation %q", alpnMismatchExpected)
	}
	for _, withCallback := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback_%v", withCallback), func(t *testing.T) {
			calls := 0
			original := &tls.Config{NextProtos: []string{"h3"}}
			if withCallback {
				original.GetConfigForClient = func(*tls.ClientHelloInfo) (*tls.Config, error) {
					calls++
					return nil, nil
				}
			}
			wrapped := setupTLSConfigForALPNMismatch(original)
			if calls != 0 {
				t.Fatal("setup called GetConfigForClient before receiving a ClientHello")
			}
			if alpnMismatchExpected == "disabled" {
				if wrapped != original {
					t.Fatalf("version %q: disabled workaround must return the original config", core.Version())
				}
				if (wrapped.GetConfigForClient != nil) != withCallback {
					t.Fatal("disabled workaround changed GetConfigForClient")
				}
				if withCallback {
					selected, err := wrapped.GetConfigForClient(&tls.ClientHelloInfo{SupportedProtos: []string{"legacy"}})
					if selected != nil || err != nil || calls != 1 {
						t.Fatalf("disabled workaround changed callback behavior: config %p, error %v, calls %d", selected, err, calls)
					}
				}
			} else if wrapped == original || wrapped.GetConfigForClient == nil {
				t.Fatalf("version %q: enabled workaround must install a callback on a clone", core.Version())
			}
			if !slices.Equal(original.NextProtos, []string{"h3"}) {
				t.Fatalf("setup changed the original ALPN list: %v", original.NextProtos)
			}
		})
	}
}

func TestSetupTLSConfigForALPNMismatchSelection(t *testing.T) {
	tests := []struct {
		name           string
		server, client []string
		want           []string
		wantClone      bool
	}{
		{name: "neither_advertises", want: nil},
		{name: "client_omits_alpn", server: []string{"h3"}, want: []string{"h3"}},
		{name: "client_empty_alpn", server: []string{"h3"}, client: []string{}, want: []string{"h3"}},
		{name: "server_omits_alpn", client: []string{"legacy", "h3"}, want: []string{"legacy"}, wantClone: true},
		{name: "server_empty_alpn", server: []string{}, client: []string{"legacy"}, want: []string{"legacy"}, wantClone: true},
		{name: "exact_match", server: []string{"h3"}, client: []string{"h3"}, want: []string{"h3"}},
		{name: "later_client_match", server: []string{"h3"}, client: []string{"legacy", "h3"}, want: []string{"h3"}},
		{name: "later_server_match", server: []string{"h3", "h2"}, client: []string{"h2"}, want: []string{"h3", "h2"}},
		{name: "later_match_on_both", server: []string{"h3", "h2"}, client: []string{"legacy", "h2"}, want: []string{"h3", "h2"}},
		{name: "preserve_server_preference", server: []string{"h3", "h2"}, client: []string{"h2", "h3"}, want: []string{"h3", "h2"}},
		{name: "single_mismatch", server: []string{"h3"}, client: []string{"legacy"}, want: []string{"legacy"}, wantClone: true},
		{name: "first_client_protocol_wins", server: []string{"h3", "h2"}, client: []string{"legacy", "other"}, want: []string{"legacy"}, wantClone: true},
		{name: "case_sensitive", server: []string{"h3"}, client: []string{"H3"}, want: []string{"H3"}, wantClone: true},
		{name: "exact_not_prefix_match", server: []string{"h3"}, client: []string{"h3-29"}, want: []string{"h3-29"}, wantClone: true},
		{name: "duplicate_client_protocols", server: []string{"h3"}, client: []string{"legacy", "legacy"}, want: []string{"legacy"}, wantClone: true},
		{name: "only_outer_config_matches", server: []string{"h3"}, client: []string{"outer-only"}, want: []string{"outer-only"}, wantClone: true},
	}
	for _, mode := range []string{"no_callback", "nil_callback_result", "original_callback_config", "selected_callback_config"} {
		t.Run(mode, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					selected := &tls.Config{NextProtos: slices.Clone(tt.server), MinVersion: tls.VersionTLS13}
					original := selected
					info := &tls.ClientHelloInfo{
						ServerName: "example.test", SupportedProtos: slices.Clone(tt.client),
						SupportedVersions: []uint16{tls.VersionTLS13},
					}
					calls := 0
					if mode == "selected_callback_config" {
						original = &tls.Config{NextProtos: []string{"outer-only"}, MinVersion: tls.VersionTLS12}
						selected.GetConfigForClient = func(*tls.ClientHelloInfo) (*tls.Config, error) {
							t.Error("selected config's callback must not be invoked recursively")
							return nil, nil
						}
					}
					if mode != "no_callback" {
						original.GetConfigForClient = func(got *tls.ClientHelloInfo) (*tls.Config, error) {
							calls++
							if got != info {
								t.Error("callback did not receive the original ClientHelloInfo")
							}
							if mode == "nil_callback_result" {
								return nil, nil
							}
							return selected, nil
						}
					}
					originalProtos := slices.Clone(original.NextProtos)
					wrapped := wrapALPNMismatchConfig(t, original)
					if calls != 0 {
						t.Fatal("setup invoked the application callback")
					}
					got, err := wrapped.GetConfigForClient(info)
					if err != nil || got == nil {
						t.Fatalf("GetConfigForClient() = (%p, %v)", got, err)
					}
					if (got != selected) != tt.wantClone {
						t.Errorf("returned config %p, selected %p; want clone %v", got, selected, tt.wantClone)
					}
					if !slices.Equal(got.NextProtos, tt.want) {
						t.Errorf("NextProtos = %v, want %v", got.NextProtos, tt.want)
					}
					if got.MinVersion != selected.MinVersion {
						t.Error("returned config lost the selected TLS minimum version")
					}
					wantCalls := 0
					if mode != "no_callback" {
						wantCalls = 1
					}
					if calls != wantCalls {
						t.Errorf("application callback called %d times, want %d", calls, wantCalls)
					}
					if tt.wantClone {
						got.NextProtos[0] = "changed-by-connection"
					}
					if !slices.Equal(selected.NextProtos, tt.server) || !slices.Equal(original.NextProtos, originalProtos) || !slices.Equal(wrapped.NextProtos, originalProtos) {
						t.Error("connection-specific ALPN changed a shared config")
					}
					if !slices.Equal(info.SupportedProtos, tt.client) {
						t.Error("connection-specific ALPN aliases the ClientHello protocol list")
					}
					if (original.GetConfigForClient != nil) != (mode != "no_callback") {
						t.Error("setup replaced the original application's callback")
					}
					if mode != "no_callback" {
						wantOriginal := selected
						if mode == "nil_callback_result" {
							wantOriginal = nil
						}
						if got, err := original.GetConfigForClient(info); got != wantOriginal || err != nil {
							t.Errorf("original callback changed: got (%p, %v), want (%p, nil)", got, err, wantOriginal)
						}
					}
				})
			}
		})
	}
}

func TestSetupTLSConfigForALPNMismatchCallbackError(t *testing.T) {
	wantErr := errors.New("application rejected ClientHello")
	for _, returnConfig := range []bool{false, true} {
		for _, protos := range [][]string{nil, {"h3"}, {"legacy"}} {
			t.Run(fmt.Sprintf("config_%v/protos_%v", returnConfig, protos), func(t *testing.T) {
				selected := &tls.Config{NextProtos: []string{"application"}}
				calls := 0
				original := &tls.Config{
					NextProtos: []string{"h3"},
					GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
						calls++
						if returnConfig {
							return selected, wantErr
						}
						return nil, wantErr
					},
				}
				wrapped := wrapALPNMismatchConfig(t, original)
				got, err := wrapped.GetConfigForClient(&tls.ClientHelloInfo{SupportedProtos: protos})
				if got != nil || err != wantErr || calls != 1 {
					t.Fatalf("GetConfigForClient() = (%p, %v), calls %d; want (nil, original error), calls 1", got, err, calls)
				}
				if !slices.Equal(original.NextProtos, []string{"h3"}) || !slices.Equal(selected.NextProtos, []string{"application"}) {
					t.Error("callback error changed a shared ALPN list")
				}
			})
		}
	}
}

func TestSetupTLSConfigForALPNMismatchPreservesTLSSettings(t *testing.T) {
	for _, withCallback := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback_%v", withCallback), func(t *testing.T) {
			certificate := &tls.Certificate{}
			verifyErr := errors.New("verification sentinel")
			selected := &tls.Config{
				NextProtos:             []string{"h3"},
				ServerName:             "selected.example",
				MinVersion:             tls.VersionTLS12,
				MaxVersion:             tls.VersionTLS13,
				ClientAuth:             tls.RequireAndVerifyClientCert,
				ClientCAs:              x509.NewCertPool(),
				RootCAs:                x509.NewCertPool(),
				SessionTicketsDisabled: true,
				CipherSuites:           []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
				CurvePreferences:       []tls.CurveID{tls.X25519},
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return certificate, nil
				},
				VerifyConnection: func(tls.ConnectionState) error { return verifyErr },
			}
			original := selected
			if withCallback {
				original = &tls.Config{GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
					return selected, nil
				}}
			}
			wrapped := wrapALPNMismatchConfig(t, original)
			got, err := wrapped.GetConfigForClient(&tls.ClientHelloInfo{SupportedProtos: []string{"legacy"}})
			if err != nil || got == nil {
				t.Fatalf("GetConfigForClient() = (%p, %v)", got, err)
			}
			if got.ServerName != selected.ServerName || got.MinVersion != selected.MinVersion || got.MaxVersion != selected.MaxVersion ||
				got.ClientAuth != selected.ClientAuth || got.ClientCAs != selected.ClientCAs || got.RootCAs != selected.RootCAs ||
				got.SessionTicketsDisabled != selected.SessionTicketsDisabled || !slices.Equal(got.CipherSuites, selected.CipherSuites) ||
				!slices.Equal(got.CurvePreferences, selected.CurvePreferences) {
				t.Error("ALPN fallback lost the selected TLS settings")
			}
			if got.GetCertificate == nil || got.VerifyConnection == nil {
				t.Fatal("ALPN fallback removed a TLS callback")
			}
			if gotCert, err := got.GetCertificate(&tls.ClientHelloInfo{}); gotCert != certificate || err != nil {
				t.Errorf("GetCertificate() = (%p, %v), want (%p, nil)", gotCert, err, certificate)
			}
			if err := got.VerifyConnection(tls.ConnectionState{}); err != verifyErr {
				t.Errorf("VerifyConnection() = %v, want original verification error", err)
			}
		})
	}
}

func TestSetupTLSConfigForALPNMismatchConcurrentConnections(t *testing.T) {
	for _, withCallback := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback_%v", withCallback), func(t *testing.T) {
			selected := &tls.Config{NextProtos: []string{"h3", "h2"}}
			original := selected
			if withCallback {
				original = &tls.Config{GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
					return selected, nil
				}}
			}
			wrapped := wrapALPNMismatchConfig(t, original)
			const connections = 64
			results := make([]*tls.Config, connections)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := range connections {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					info := &tls.ClientHelloInfo{SupportedProtos: []string{fmt.Sprintf("legacy-%d", i)}}
					got, err := wrapped.GetConfigForClient(info)
					if err != nil || got == nil {
						t.Errorf("connection %d: config %p, error %v", i, got, err)
						return
					}
					results[i] = got
					// A subsequent handshake must still use the advertised server ALPN.
					matched, err := wrapped.GetConfigForClient(&tls.ClientHelloInfo{SupportedProtos: []string{"h2", "h3"}})
					if err != nil || matched != selected {
						t.Errorf("connection %d: matching ALPN returned (%p, %v), want original selected config %p", i, matched, err, selected)
					}
				}()
			}
			close(start)
			wg.Wait()
			seen := make(map[*tls.Config]bool)
			for i, got := range results {
				if got == nil {
					continue
				}
				if got == selected || got == wrapped || seen[got] {
					t.Errorf("connection %d reused a shared config", i)
				}
				seen[got] = true
				if want := []string{fmt.Sprintf("legacy-%d", i)}; !slices.Equal(got.NextProtos, want) {
					t.Errorf("connection %d: NextProtos = %v, want %v", i, got.NextProtos, want)
				}
				// Also detect separate config structs sharing the same fallback slice.
				if len(got.NextProtos) > 0 {
					got.NextProtos[0] = "changed-by-another-connection"
				}
			}
			if !slices.Equal(selected.NextProtos, []string{"h3", "h2"}) {
				t.Errorf("concurrent handshakes changed shared NextProtos: %v", selected.NextProtos)
			}
		})
	}
}

func TestSetupTLSConfigForALPNMismatchHandshake(t *testing.T) {
	server, client := alpnMismatchTLSConfigs(t)
	tests := []struct {
		name           string
		server, client []string
		want           string
	}{
		{name: "mismatch", server: []string{"h3"}, client: []string{"legacy", "other"}, want: "legacy"},
		{name: "server_preference", server: []string{"h3", "h2"}, client: []string{"h2", "h3"}, want: "h3"},
		{name: "later_match", server: []string{"h3"}, client: []string{"legacy", "h3"}, want: "h3"},
		{name: "client_omits_alpn", server: []string{"h3"}},
		{name: "server_omits_alpn", client: []string{"legacy"}, want: "legacy"},
		{name: "neither_advertises"},
	}
	for _, withCallback := range []bool{false, true} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("callback_%v/%s", withCallback, tt.name), func(t *testing.T) {
				selected := server.Clone()
				selected.NextProtos = slices.Clone(tt.server)
				original := selected
				calls := 0
				if withCallback {
					// Only the selected config has a certificate; a successful handshake
					// therefore also proves the application config is honored.
					original = &tls.Config{GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
						calls++
						if info.ServerName != client.ServerName || !slices.Equal(info.SupportedProtos, tt.client) {
							t.Error("application callback received unexpected ClientHello fields")
						}
						return selected, nil
					}}
				}
				clientConf := client.Clone()
				clientConf.NextProtos = slices.Clone(tt.client)
				serverState, clientState := alpnMismatchHandshake(t, wrapALPNMismatchConfig(t, original), clientConf)
				if serverState.NegotiatedProtocol != tt.want || clientState.NegotiatedProtocol != tt.want {
					t.Errorf("negotiated ALPN: server %q, client %q; want %q", serverState.NegotiatedProtocol, clientState.NegotiatedProtocol, tt.want)
				}
				if withCallback && calls != 1 {
					t.Errorf("application callback called %d times, want 1", calls)
				}
				if !slices.Equal(selected.NextProtos, tt.server) {
					t.Error("TLS handshake mutated the shared ALPN list")
				}
			})
		}
	}
}

func TestSetupTLSConfigForALPNMismatchSessionResumption(t *testing.T) {
	for _, ticketMode := range []string{"automatic", "explicit", "disabled"} {
		t.Run(ticketMode, func(t *testing.T) {
			server, client := alpnMismatchTLSConfigs(t)
			server.NextProtos = []string{"h3"}
			if ticketMode == "explicit" {
				server.SetSessionTicketKeys([][32]byte{{1, 2, 3}})
			}
			server.SessionTicketsDisabled = ticketMode == "disabled"
			client.NextProtos = []string{"legacy"}
			client.ClientSessionCache = tls.NewLRUClientSessionCache(1)
			// Wrap the same cold config twice before either clone handles a
			// connection. Initializing keys after cloning breaks cross-clone reuse.
			first := wrapALPNMismatchConfig(t, server)
			second := wrapALPNMismatchConfig(t, server)
			for i, conf := range []*tls.Config{first, second, first} {
				serverState, clientState := alpnMismatchHandshake(t, conf, client)
				wantResume := i > 0 && ticketMode != "disabled"
				if serverState.DidResume != wantResume || clientState.DidResume != wantResume {
					t.Errorf("connection %d: resumed server %v, client %v; want %v", i, serverState.DidResume, clientState.DidResume, wantResume)
				}
				if serverState.NegotiatedProtocol != "legacy" || clientState.NegotiatedProtocol != "legacy" {
					t.Errorf("connection %d lost fallback ALPN: server %q, client %q", i, serverState.NegotiatedProtocol, clientState.NegotiatedProtocol)
				}
			}
			if !slices.Equal(server.NextProtos, []string{"h3"}) {
				t.Error("session resumption changed the shared ALPN list")
			}
		})
	}
}

func wrapALPNMismatchConfig(t *testing.T, original *tls.Config) *tls.Config {
	t.Helper()
	wrapped := setupTLSConfigForALPNMismatch(original)
	if wrapped == original || wrapped.GetConfigForClient == nil {
		t.Fatalf("version %q: expected ALPN workaround on a cloned TLS config", core.Version())
	}
	return wrapped
}

func alpnMismatchTLSConfigs(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()
	generated, err := cert.Generate(nil, cert.DNSNames("example.test"))
	if err != nil {
		t.Fatal(err)
	}
	certPEM, keyPEM := generated.ToPEM()
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certPEM) {
		t.Fatal("could not add test certificate to client roots")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
	}, &tls.Config{
		ServerName: "example.test", RootCAs: roots, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
	}
}

func alpnMismatchHandshake(t *testing.T, serverConfig, clientConfig *tls.Config) (tls.ConnectionState, tls.ConnectionState) {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	if err := serverConn.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	server := tls.Server(serverConn, serverConfig)
	client := tls.Client(clientConn, clientConfig)
	type result struct {
		state tls.ConnectionState
		err   error
	}
	serverDone := make(chan result, 1)
	go func() {
		err := server.HandshakeContext(ctx)
		if err == nil {
			// Reading application data makes the client consume TLS 1.3 session
			// tickets before the next connection attempts resumption.
			_, err = server.Write([]byte{1})
		}
		serverDone <- result{state: server.ConnectionState(), err: err}
	}()
	clientErr := client.HandshakeContext(ctx)
	if clientErr == nil {
		var data [1]byte
		_, clientErr = io.ReadFull(client, data[:])
	}
	if clientErr != nil {
		// Unblock the server immediately on failure instead of waiting for its deadline.
		clientConn.Close()
	}
	serverResult := <-serverDone
	if clientErr != nil || serverResult.err != nil {
		t.Fatalf("TLS handshake/data exchange failed: server %v, client %v", serverResult.err, clientErr)
	}
	return serverResult.state, client.ConnectionState()
}
