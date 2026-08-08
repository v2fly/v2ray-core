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

func newTransportContext(t *testing.T) context.Context {
	t.Helper()

	ctx := context.Background()
	defaultNetworkImpl := systemnetworkimpl.NewSystemNetworkDefault()
	rootEnv := environment.NewRootEnvImpl(
		ctx,
		transientstorageimpl.NewScopedTransientStorageImpl(),
		defaultNetworkImpl.Dialer(),
		defaultNetworkImpl.Listener(),
		filesystemimpl.NewDefaultFileSystemDefaultImpl(),
		deferredpersistentstorage.NewDeferredPersistentStorage(ctx),
	)
	transportEnvironment, err := rootEnv.ProxyEnvironment(protocolName).NarrowScopeToTransport(protocolName)
	if err != nil {
		t.Fatal(err)
	}
	return envctx.ContextWithEnvironment(ctx, transportEnvironment)
}

// testEarlyDataRoundTrip checks that everything written by the client reaches the
// server exactly once, no matter how the payload is split between the early data
// header and the connection itself.
func testEarlyDataRoundTrip(t *testing.T, maxEarlyData int32, payloadSize int) {
	t.Helper()

	config := &Config{
		Path:                "/httpupgrade",
		MaxEarlyData:        maxEarlyData,
		EarlyDataHeaderName: "Sec-WebSocket-Key",
	}
	streamSettings := &internet.MemoryStreamConfig{
		ProtocolName:     protocolName,
		ProtocolSettings: config,
	}

	payload := make([]byte, payloadSize)
	common.Must2(rand.Read(payload))

	ctx := newTransportContext(t)

	type result struct {
		payload []byte
		err     error
	}
	received := make(chan result, 1)

	port := tcp.PickPort()
	listener, err := listenHTTPUpgrade(ctx, net.LocalHostIP, port, streamSettings, func(conn internet.Connection) {
		go func(c internet.Connection) {
			defer c.Close()

			b := make([]byte, payloadSize)
			if _, err := io.ReadFull(c, b); err != nil {
				received <- result{err: newError("failed to read the whole payload").Base(err)}
				return
			}
			// The client wrote the payload once, so nothing else may follow it.
			if err := c.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				received <- result{err: err}
				return
			}
			var extra [1]byte
			if n, _ := c.Read(extra[:]); n > 0 {
				received <- result{err: newError("received more data than the client has written")}
				return
			}
			received <- result{payload: b}
		}(conn)
	})
	common.Must(err)
	defer listener.Close()

	conn, err := dialhttpUpgrade(ctx, net.TCPDestination(net.LocalHostIP, port), streamSettings)
	common.Must(err)
	defer conn.Close()

	common.Must2(conn.Write(payload))

	select {
	case r := <-received:
		if r.err != nil {
			t.Fatal(r.err)
		}
		if !bytes.Equal(r.payload, payload) {
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

// TestNegativeMaxEarlyData makes sure a nonsensical MaxEarlyData does not take down
// the whole instance: the payload is simply sent without using the early data header.
func TestNegativeMaxEarlyData(t *testing.T) {
	testEarlyDataRoundTrip(t, -1, 4096)
}

// readerEndingWithEOF returns the last chunk of its payload together with io.EOF,
// which io.Reader implementations (io.MultiReader among them) are allowed to do.
type readerEndingWithEOF struct {
	payload []byte
}

func (r *readerEndingWithEOF) Read(p []byte) (int, error) {
	n := copy(p, r.payload)
	r.payload = r.payload[n:]
	if len(r.payload) == 0 {
		return n, io.EOF
	}
	return n, nil
}

// TestPendingReadEndingWithEOF makes sure the pending read is not dropped when the
// reader reports io.EOF alongside the last bytes it returns.
func TestPendingReadEndingWithEOF(t *testing.T) {
	clientConn, serverConn := gonet.Pipe()
	defer clientConn.Close()

	go func() {
		clientConn.Write([]byte("world")) // nolint: errcheck
	}()

	conn := newConnectionWithPendingRead(serverConn, serverConn.RemoteAddr(), &readerEndingWithEOF{payload: []byte("hello")})
	defer conn.Close()
	common.Must(conn.SetReadDeadline(time.Now().Add(time.Second * 10)))

	received := make([]byte, len("helloworld"))
	if _, err := io.ReadFull(conn, received); err != nil {
		t.Fatal("failed to read the whole payload: ", err)
	}
	if string(received) != "helloworld" {
		t.Error("unexpected payload: ", string(received))
	}
}

// TestEarlyDataArrivingWithRequest covers the case where the part of the early data
// that did not fit into the header reaches the server together with the upgrade
// request, and is therefore consumed by the buffered reader parsing that request.
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

	// net.Pipe hands the whole buffer to the server in a single read.
	clientConn, serverConn := gonet.Pipe()
	defer clientConn.Close()

	go func() {
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
