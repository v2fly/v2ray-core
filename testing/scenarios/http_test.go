package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
	"github.com/ghxhy/v2ray-core/v5/proxy/freedom"
	v2http "github.com/ghxhy/v2ray-core/v5/proxy/http"
	v2httptest "github.com/ghxhy/v2ray-core/v5/testing/servers/http"
	"github.com/ghxhy/v2ray-core/v5/testing/servers/tcp"
)

func TestHttpConformance(t *testing.T) {
	httpServerPort := tcp.PickPort()
	httpServer := &v2httptest.Server{
		Port:        httpServerPort,
		PathHandler: make(map[string]http.HandlerFunc),
	}
	_, err := httpServer.Start()
	common.Must(err)
	defer httpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://127.0.0.1:" + httpServerPort.String())
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("status: ", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		common.Must(err)
		if string(content) != "Home" {
			t.Fatal("body: ", string(content))
		}
	}
}

func TestHttpError(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: func(msg []byte) []byte {
			return []byte{}
		},
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	time.AfterFunc(time.Second*2, func() {
		tcpServer.ShouldClose = true
	})

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://127.0.0.1:" + dest.Port.String())
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 503 {
			t.Error("status: ", resp.StatusCode)
		}
	}
}

func TestHTTPConnectMethod(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		payload := make([]byte, 1024*64)
		common.Must2(rand.Read(payload))

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, "Connect", "http://"+dest.NetAddr()+"/", bytes.NewReader(payload))
		req.Header.Set("X-a", "b")
		req.Header.Set("X-b", "d")
		common.Must(err)

		resp, err := client.Do(req)
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("status: ", resp.StatusCode)
		}

		content := make([]byte, len(payload))
		common.Must2(io.ReadFull(resp.Body, content))
		if r := cmp.Diff(content, xor(payload)); r != "" {
			t.Fatal(r)
		}
	}
}

func TestHttpPost(t *testing.T) {
	httpServerPort := tcp.PickPort()
	httpServer := &v2httptest.Server{
		Port: httpServerPort,
		PathHandler: map[string]http.HandlerFunc{
			"/testpost": func(w http.ResponseWriter, r *http.Request) {
				payload, err := buf.ReadAllToBytes(r.Body)
				r.Body.Close()

				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte("Unable to read all payload"))
					return
				}
				payload = xor(payload)
				w.Write(payload)
			},
		},
	}

	_, err := httpServer.Start()
	common.Must(err)
	defer httpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		payload := make([]byte, 1024*64)
		common.Must2(rand.Read(payload))

		resp, err := client.Post("http://127.0.0.1:"+httpServerPort.String()+"/testpost", "application/x-www-form-urlencoded", bytes.NewReader(payload))
		common.Must(err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("status: ", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		common.Must(err)
		if r := cmp.Diff(content, xor(payload)); r != "" {
			t.Fatal(r)
		}
	}
}

func setProxyBasicAuth(req *http.Request, user, pass string) {
	req.SetBasicAuth(user, pass)
	req.Header.Set("Proxy-Authorization", req.Header.Get("Authorization"))
	req.Header.Del("Authorization")
}

func TestHttpBasicAuth(t *testing.T) {
	httpServerPort := tcp.PickPort()
	httpServer := &v2httptest.Server{
		Port:        httpServerPort,
		PathHandler: make(map[string]http.HandlerFunc),
	}
	_, err := httpServer.Start()
	common.Must(err)
	defer httpServer.Close()

	serverPort := tcp.PickPort()
	serverConfig := &core.Config{
		Inbound: []*core.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{
					Accounts: map[string]string{
						"a": "b",
					},
				}),
			},
		},
		Outbound: []*core.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	common.Must(err)
	defer CloseAllServers(servers)

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		{
			resp, err := client.Get("http://127.0.0.1:" + httpServerPort.String())
			common.Must(err)
			defer resp.Body.Close()
			if resp.StatusCode != 407 {
				t.Fatal("status: ", resp.StatusCode)
			}
		}

		{
			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:"+httpServerPort.String(), nil)
			common.Must(err)

			setProxyBasicAuth(req, "a", "c")
			resp, err := client.Do(req)
			common.Must(err)
			defer resp.Body.Close()
			if resp.StatusCode != 407 {
				t.Fatal("status: ", resp.StatusCode)
			}
		}

		{
			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:"+httpServerPort.String(), nil)
			common.Must(err)

			setProxyBasicAuth(req, "a", "b")
			resp, err := client.Do(req)
			common.Must(err)
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Fatal("status: ", resp.StatusCode)
			}

			content, err := io.ReadAll(resp.Body)
			common.Must(err)
			if string(content) != "Home" {
				t.Fatal("body: ", string(content))
			}
		}
	}
}
