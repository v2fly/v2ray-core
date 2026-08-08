package httpupgrade

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	gonet "net"
	"net/http"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/deferredpersistentstorage"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/environment/filesystemimpl"
	"github.com/v2fly/v2ray-core/v5/common/environment/systemnetworkimpl"
	"github.com/v2fly/v2ray-core/v5/common/environment/transientstorageimpl"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func newTransportEnvironment(t *testing.T, ctx context.Context) environment.TransportEnvironment {
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
	proxyEnvironment := rootEnv.ProxyEnvironment(protocolName)
	transportEnvironment, err := proxyEnvironment.NarrowScopeToTransport(protocolName)
	if err != nil {
		t.Fatal(err)
	}
	return transportEnvironment
}

// testEarlyDataRoundTrip verifies that everything written by the client is delivered
// to the server exactly once, no matter how the payload is split between the early
// data header and the connection itself.
func testEarlyDataRoundTrip(t *testing.T, maxEarlyData int32, payloadSize int) {
	t.Helper()

	config := &Config{
		Path:                "/httpupgrade",
		MaxEarlyData:        maxEarlyData,
		EarlyDataHeaderName: "Sec-WebSocket-Key",
	}

	payload := make([]byte, payloadSize)
	common.Must2(rand.Read(payload))

	ctx := envctx.ContextWithEnvironment(context.Background(), newTransportEnvironment(t, context.Background()))

	received := make(chan []byte, 1)
	port := tcp.PickPort()
	listener, err := listenHTTPUpgrade(ctx, net.LocalHostIP, port, &internet.MemoryStreamConfig{
		ProtocolName:     protocolName,
		ProtocolSettings: config,
	}, func(conn internet.Connection) {
		go func(c internet.Connection) {
			defer c.Close()

			b := make([]byte, payloadSize)
			if _, err := io.ReadFull(c, b); err != nil {
				received <- nil
				return
			}
			// Nothing else is supposed to be on the wire.
			common.Must(c.SetReadDeadline(time.Now().Add(time.Second)))
			var extra [1]byte
			if n, _ := c.Read(extra[:]); n > 0 {
				received <- nil
				return
			}
			received <- b
		}(conn)
	})
	common.Must(err)
	defer listener.Close()

	conn, err := dialhttpUpgrade(ctx, net.TCPDestination(net.LocalHostIP, port), &internet.MemoryStreamConfig{
		ProtocolName:     protocolName,
		ProtocolSettings: config,
	})
	common.Must(err)
	defer conn.Close()

	_, err = conn.Write(payload)
	common.Must(err)

	select {
	case b := <-received:
		if !bytes.Equal(b, payload) {
			t.Error("payload received by the server does not match the payload sent by the client")
		}
	case <-time.After(time.Second * 10):
		t.Error("timeout waiting for the server to receive the payload")
	}
}

func TestEarlyDataLargerThanMaxEarlyData(t *testing.T) {
	testEarlyDataRoundTrip(t, 32, 4096)
}

func TestEarlyDataSmallerThanMaxEarlyData(t *testing.T) {
	testEarlyDataRoundTrip(t, 2048, 1024)
}

func TestWithoutEarlyData(t *testing.T) {
	testEarlyDataRoundTrip(t, 0, 4096)
}

// TestEarlyDataArrivingWithRequest covers the case where the remainder of the early
// data reaches the server in the same read as the upgrade request itself, and would
// therefore be swallowed by the buffered reader used to parse the request.
func TestEarlyDataArrivingWithRequest(t *testing.T) {
	config := &Config{
		Path:                "/httpupgrade",
		MaxEarlyData:        32,
		EarlyDataHeaderName: "Sec-WebSocket-Key",
	}

	payload := make([]byte, 1024)
	common.Must2(rand.Read(payload))

	req, err := http.NewRequest("GET", config.GetNormalizedPath(), nil)
	common.Must(err)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set(config.EarlyDataHeaderName, base64.URLEncoding.EncodeToString(payload[:config.MaxEarlyData]))

	var request bytes.Buffer
	common.Must(req.Write(&request))
	common.Must2(request.Write(payload[config.MaxEarlyData:]))

	clientConn, serverConn := gonet.Pipe()
	defer clientConn.Close()

	go func() {
		defer clientConn.Close()
		if _, err := clientConn.Write(request.Bytes()); err != nil {
			return
		}
		io.Copy(io.Discard, clientConn) // nolint: errcheck
	}()

	upgradedConn, err := (&server{config: config}).upgrade(serverConn)
	common.Must(err)
	defer upgradedConn.Close()

	common.Must(upgradedConn.SetReadDeadline(time.Now().Add(time.Second * 10)))
	received := make([]byte, len(payload))
	if _, err := io.ReadFull(upgradedConn, received); err != nil {
		t.Fatal("failed to read the whole payload: ", err)
	}
	if !bytes.Equal(received, payload) {
		t.Error("payload received by the server does not match the payload sent by the client")
	}
}
