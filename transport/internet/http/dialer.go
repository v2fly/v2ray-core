// +build !confonly

package http

import (
	"context"
	gotls "crypto/tls"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/http2"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/tls"
	"github.com/v2fly/v2ray-core/v4/transport/pipe"
)

var (
	globalDialerMap    map[net.Destination]*http.Client
	globalDialerAccess sync.Mutex
)

func getHTTPClient(ctx context.Context, dest net.Destination, tlsSettings *tls.Config) *http.Client {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	if globalDialerMap == nil {
		globalDialerMap = make(map[net.Destination]*http.Client)
	}

	if client, found := globalDialerMap[dest]; found {
		return client
	}

	transport := &http2.Transport{
		DialTLS: func(network string, addr string, tlsConfig *gotls.Config) (net.Conn, error) {
			rawHost, rawPort, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			if len(rawPort) == 0 {
				rawPort = "443"
			}
			port, err := net.PortFromString(rawPort)
			if err != nil {
				return nil, err
			}
			address := net.ParseAddress(rawHost)

			pconn, err := internet.DialSystem(ctx, net.TCPDestination(address, port), nil)
			if err != nil {
				return nil, err
			}

			cn := gotls.Client(pconn, tlsConfig)
			if err := cn.Handshake(); err != nil {
				return nil, err
			}
			if !tlsConfig.InsecureSkipVerify {
				if err := cn.VerifyHostname(tlsConfig.ServerName); err != nil {
					return nil, err
				}
			}
			state := cn.ConnectionState()
			if p := state.NegotiatedProtocol; p != http2.NextProtoTLS {
				return nil, newError("http2: unexpected ALPN protocol " + p + "; want q" + http2.NextProtoTLS).AtError()
			}
			return cn, nil
		},
		TLSClientConfig: tlsSettings.GetTLSConfig(tls.WithDestination(dest)),
	}

	client := &http.Client{
		Transport: transport,
	}

	globalDialerMap[dest] = client
	return client
}

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	httpSettings := streamSettings.ProtocolSettings.(*Config)
	tlsConfig := tls.ConfigFromStreamSettings(streamSettings)
	if tlsConfig == nil {
		return nil, newError("TLS must be enabled for http transport.").AtWarning()
	}
	client := getHTTPClient(ctx, dest, tlsConfig)

	opts := pipe.OptionsFromContext(ctx)
	preader, pwriter := pipe.New(opts...)
	breader := &buf.BufferedReader{Reader: preader}

	httpMethod := "PUT"
	if httpSettings.Method != "" {
		httpMethod = httpSettings.Method
	}

	httpHeaders := make(http.Header)

	for _, httpHeader := range httpSettings.Header {
		for _, httpHeaderValue := range httpHeader.Value {
			httpHeaders.Set(httpHeader.Name, httpHeaderValue)
		}
	}

	request := &http.Request{
		Method: httpMethod,
		Host:   httpSettings.getRandomHost(),
		Body:   breader,
		URL: &url.URL{
			Scheme: "https",
			Host:   dest.NetAddr(),
			Path:   httpSettings.getNormalizedPath(),
		},
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		ProtoMinor: 0,
		Header:     httpHeaders,
	}
	// Disable any compression method from server.
	request.Header.Set("Accept-Encoding", "identity")

	response, err := client.Do(request) // nolint: bodyclose
	if err != nil {
		return nil, newError("failed to dial to ", dest).Base(err).AtWarning()
	}
	if response.StatusCode != 200 {
		return nil, newError("unexpected status", response.StatusCode).AtWarning()
	}

	bwriter := buf.NewBufferedWriter(pwriter)
	common.Must(bwriter.SetBuffered(false))
	return net.NewConnection(
		net.ConnectionOutput(response.Body),
		net.ConnectionInput(bwriter),
		net.ConnectionOnClose(common.ChainedClosable{breader, bwriter, response.Body}),
	), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
