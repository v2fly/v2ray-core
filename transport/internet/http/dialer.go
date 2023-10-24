package http

import (
	"context"
	gotls "crypto/tls"
	gonet "net"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/http2"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

var (
	globalDialerMap    map[net.Destination]*http.Client
	globalDialerAccess sync.Mutex
)

type dialerCanceller func()

func getHTTPClient(ctx context.Context, dest net.Destination, securityEngine *security.Engine, streamSettings *internet.MemoryStreamConfig) (*http.Client, dialerCanceller) {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	canceller := func() {
		globalDialerAccess.Lock()
		defer globalDialerAccess.Unlock()
		delete(globalDialerMap, dest)
	}

	if globalDialerMap == nil {
		globalDialerMap = make(map[net.Destination]*http.Client)
	}

	if client, found := globalDialerMap[dest]; found {
		return client, canceller
	}

	transport := &http2.Transport{
		DialTLSContext: func(_ context.Context, network, addr string, tlsConfig *gotls.Config) (gonet.Conn, error) {
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

			detachedContext := core.ToBackgroundDetachedContext(ctx)
			pconn, err := internet.DialSystem(detachedContext, net.TCPDestination(address, port), streamSettings.SocketSettings)
			if err != nil {
				return nil, err
			}

			cn, err := (*securityEngine).Client(pconn,
				security.OptionWithDestination{Dest: dest})
			if err != nil {
				return nil, err
			}

			protocol := ""
			if connAPLNGetter, ok := cn.(security.ConnectionApplicationProtocol); ok {
				connectionALPN, err := connAPLNGetter.GetConnectionApplicationProtocol()
				if err != nil {
					return nil, newError("failed to get connection ALPN").Base(err).AtWarning()
				}
				protocol = connectionALPN
			}

			if protocol != http2.NextProtoTLS {
				return nil, newError("http2: unexpected ALPN protocol " + protocol + "; want q" + http2.NextProtoTLS).AtError()
			}
			return cn, nil
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	globalDialerMap[dest] = client
	return client, canceller
}

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	httpSettings := streamSettings.ProtocolSettings.(*Config)
	securityEngine, err := security.CreateSecurityEngineFromSettings(ctx, streamSettings)
	if err != nil {
		return nil, newError("unable to create security engine").Base(err)
	}
	if securityEngine == nil {
		return nil, newError("TLS must be enabled for http transport.").AtWarning()
	}
	client, canceller := getHTTPClient(ctx, dest, &securityEngine, streamSettings)

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
		canceller()
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
