package dtls_test

import (
	"bytes"
	"context"
	"io"
	gonet "net"
	"testing"
	"time"

	piondtls "github.com/pion/dtls/v3"
	piondtlsnet "github.com/pion/dtls/v3/pkg/net"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/deferredpersistentstorage"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/environment/filesystemimpl"
	"github.com/v2fly/v2ray-core/v5/common/environment/systemnetworkimpl"
	"github.com/v2fly/v2ray-core/v5/common/environment/transientstorageimpl"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transportdtls "github.com/v2fly/v2ray-core/v5/transport/internet/dtls"
)

func TestListenerExportsClientIdentity(t *testing.T) {
	ctx := context.Background()
	ctx = envctx.ContextWithEnvironment(ctx, newTransportEnvironment(t, ctx, "dtls"))

	accepted := make(chan internet.Connection, 1)
	listener, err := transportdtls.ListenDTLS(ctx, v2net.LocalHostIP, 0, &internet.MemoryStreamConfig{
		ProtocolName: "dtls",
		ProtocolSettings: &transportdtls.Config{
			Mode:                   transportdtls.DTLSMode_PSK,
			Psk:                    []byte("shared-secret"),
			Mtu:                    1200,
			ReplayProtectionWindow: 64,
		},
	}, func(conn internet.Connection) {
		accepted <- conn
	})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverAddr, ok := listener.Addr().(*gonet.UDPAddr)
	if !ok {
		t.Fatalf("unexpected listener addr type %T", listener.Addr())
	}

	rawConn, err := gonet.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer rawConn.Close()

	identity := []byte("rrpit-client-identity")
	clientConn, err := piondtls.Client(piondtlsnet.PacketConnFromConn(rawConn), rawConn.RemoteAddr(), &piondtls.Config{
		MTU:                    1200,
		ReplayProtectionWindow: 64,
		PSK: func([]byte) ([]byte, error) {
			return []byte("shared-secret"), nil
		},
		PSKIdentityHint: identity,
		CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	handshakeErr := make(chan error, 1)
	go func() {
		handshakeErr <- clientConn.Handshake()
	}()

	var serverConn internet.Connection
	select {
	case serverConn = <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for accepted DTLS connection")
	}
	defer serverConn.Close()

	if got := transportdtls.ClientIdentity(serverConn); !bytes.Equal(got, identity) {
		t.Fatalf("unexpected client identity: got %q want %q", got, identity)
	}

	select {
	case err := <-handshakeErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for client handshake")
	}

	serverErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 4)
		if _, err := io.ReadFull(serverConn, buf); err != nil {
			serverErr <- err
			return
		}
		if !bytes.Equal(buf, []byte("ping")) {
			serverErr <- io.ErrUnexpectedEOF
			return
		}
		_, err := serverConn.Write([]byte("pong"))
		serverErr <- err
	}()

	if _, err := clientConn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}

	reply := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, reply); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(reply, []byte("pong")) {
		t.Fatalf("unexpected reply: got %q want %q", reply, []byte("pong"))
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server echo")
	}
}

func newTransportEnvironment(t *testing.T, ctx context.Context, scope string) environment.TransportEnvironment {
	t.Helper()

	defaultNetworkImpl := systemnetworkimpl.NewSystemNetworkDefault()
	defaultFilesystemImpl := filesystemimpl.NewDefaultFileSystemDefaultImpl()
	deferredPersistentStorageImpl := deferredpersistentstorage.NewDeferredPersistentStorage(ctx)
	rootEnv := environment.NewRootEnvImpl(
		ctx,
		transientstorageimpl.NewScopedTransientStorageImpl(),
		defaultNetworkImpl.Dialer(),
		defaultNetworkImpl.Listener(),
		defaultFilesystemImpl,
		deferredPersistentStorageImpl,
	)
	proxyEnvironment := rootEnv.ProxyEnvironment(scope)
	transportEnvironment, err := proxyEnvironment.NarrowScopeToTransport(scope)
	if err != nil {
		t.Fatal(err)
	}
	return transportEnvironment
}
