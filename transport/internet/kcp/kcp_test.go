package kcp_test

import (
	"context"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/environment/systemnetworkimpl"
	"github.com/v2fly/v2ray-core/v5/common/environment/transientstorageimpl"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	. "github.com/v2fly/v2ray-core/v5/transport/internet/kcp"
)

func TestDialAndListen(t *testing.T) {
	ctx := context.Background()
	defaultNetworkImpl := systemnetworkimpl.NewSystemNetworkDefault()
	rootEnv := environment.NewRootEnvImpl(ctx, transientstorageimpl.NewScopedTransientStorageImpl(), defaultNetworkImpl.Dialer(), defaultNetworkImpl.Listener())
	proxyEnvironment := rootEnv.ProxyEnvironment("o")
	transportEnvironment, err := proxyEnvironment.NarrowScopeToTransport("kcp")
	if err != nil {
		t.Fatal(err)
	}
	ctx = envctx.ContextWithEnvironment(ctx, transportEnvironment)

	listener, err := NewListener(ctx, net.LocalHostIP, net.Port(0), &internet.MemoryStreamConfig{
		ProtocolName:     "mkcp",
		ProtocolSettings: &Config{},
	}, func(conn internet.Connection) {
		go func(c internet.Connection) {
			payload := make([]byte, 4096)
			for {
				nBytes, err := c.Read(payload)
				if err != nil {
					break
				}
				for idx, b := range payload[:nBytes] {
					payload[idx] = b ^ 'c'
				}
				c.Write(payload[:nBytes])
			}
			c.Close()
		}(conn)
	})
	common.Must(err)
	defer listener.Close()

	port := net.Port(listener.Addr().(*net.UDPAddr).Port)

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(func() error {
			clientConn, err := DialKCP(ctx, net.UDPDestination(net.LocalHostIP, port), &internet.MemoryStreamConfig{
				ProtocolName:     "mkcp",
				ProtocolSettings: &Config{},
			})
			if err != nil {
				return err
			}
			defer clientConn.Close()

			clientSend := make([]byte, 1024*1024)
			rand.Read(clientSend)
			go clientConn.Write(clientSend)

			clientReceived := make([]byte, 1024*1024)
			common.Must2(io.ReadFull(clientConn, clientReceived))

			clientExpected := make([]byte, 1024*1024)
			for idx, b := range clientSend {
				clientExpected[idx] = b ^ 'c'
			}
			if r := cmp.Diff(clientReceived, clientExpected); r != "" {
				return errors.New(r)
			}
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 60 && listener.ActiveConnections() > 0; i++ {
		time.Sleep(500 * time.Millisecond)
	}
	if v := listener.ActiveConnections(); v != 0 {
		t.Error("active connections: ", v)
	}
}
