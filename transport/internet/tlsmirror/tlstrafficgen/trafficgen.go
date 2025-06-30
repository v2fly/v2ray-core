package tlstrafficgen

import (
	"bufio"
	"context"
	"io"
	"math/big"
	"net/http"
	"net/url"

	cryptoRand "crypto/rand"
	"golang.org/x/net/http2"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type TrafficGenerator struct {
	config *Config
	ctx    context.Context

	destination net.Destination
	tag         string
}

func NewTrafficGenerator(ctx context.Context, config *Config) *TrafficGenerator {
	return &TrafficGenerator{
		ctx:    ctx,
		config: config,
	}
}

type trafficGeneratorManagedConnectionController struct {
	readyCtx  context.Context
	readyDone context.CancelFunc

	recallCtx  context.Context
	recallDone context.CancelFunc
}

func (t *trafficGeneratorManagedConnectionController) WaitConnectionReady() context.Context {
	return t.readyCtx
}

func (t *trafficGeneratorManagedConnectionController) RecallTrafficGenerator() error {
	t.recallDone()
	return nil
}

func (generator *TrafficGenerator) GenerateNextTraffic(ctx context.Context) error {
	transportEnvironment := envctx.EnvironmentFromContext(generator.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	carrierConnectionReadyCtx, carrierConnectionReadyDone := context.WithCancel(generator.ctx)
	carrierConnectionRecallCtx, carrierConnectionRecallDone := context.WithCancel(generator.ctx)

	trafficController := &trafficGeneratorManagedConnectionController{
		readyCtx:   carrierConnectionReadyCtx,
		readyDone:  carrierConnectionReadyDone,
		recallCtx:  carrierConnectionRecallCtx,
		recallDone: carrierConnectionRecallDone,
	}

	var trafficControllerIfce tlsmirror.TrafficGeneratorManagedConnection = trafficController
	managedConnectionContextValue := context.WithValue(generator.ctx,
		tlsmirror.TrafficGeneratorManagedConnectionContextKey, trafficControllerIfce)

	conn, err := dialer(managedConnectionContextValue, generator.destination, generator.tag)
	if err != nil {
		return newError("failed to dial to destination").Base(err).AtWarning()
	}
	tlsConn, err := generator.tlsHandshake(conn)
	if err != nil {
		return newError("failed to create TLS connection").Base(err).AtWarning()
	}
	getAlpn, ok := tlsConn.(security.ConnectionApplicationProtocol)
	if !ok {
		return newError("TLS connection does not support getting ALPN").AtWarning()
	}
	alpn, err := getAlpn.GetConnectionApplicationProtocol()
	if err != nil {
		return newError("failed to get ALPN from TLS connection").Base(err).AtWarning()
	}
	steps := generator.config.Steps
	currentStep := 0
	httpRoundTripper, err := newSingleConnectionHTTPTransport(tlsConn, alpn)
	if err != nil {
		return newError("failed to create HTTP transport").Base(err).AtWarning()
	}
	for {
		if currentStep >= len(steps) {
			return tlsConn.Close()
		}

		step := steps[currentStep]

		url := url.URL{
			Scheme: "https",
			Host:   step.Host,
			Path:   step.Path,
		}

		httpReq := &http.Request{Host: url.Hostname(), Method: step.Method, URL: &url}

		if step.Headers != nil {
			header := make(http.Header, len(step.Headers))
			for _, v := range step.Headers {
				if v.Value != "" {
					header.Add(v.Name, v.Value)
				}
				if v.Values != nil {
					for _, value := range v.Values {
						header.Add(v.Name, value)
					}
				}
			}
			httpReq.Header = header
		}

		resp, err := httpRoundTripper.RoundTrip(httpReq)
		if err != nil {
			return newError("failed to send HTTP request").Base(err).AtWarning()
		}
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			return newError("failed to read HTTP response body").Base(err).AtWarning()
		}
		err = resp.Body.Close()
		if err != nil {
			return newError("failed to close HTTP response body").Base(err).AtWarning()
		}

		if step.ConnectionReady {
			carrierConnectionReadyDone()
		}

		if step.ConnectionRecallExit {
			if carrierConnectionRecallCtx.Err() != nil {
				return tlsConn.Close()
			}
		}

		if step.NextStep == nil {
			currentStep++
		} else {
			overallWeight := int32(0)
			for _, nextStep := range step.NextStep {
				overallWeight += nextStep.Weight
			}
			maxBound := big.NewInt(int64(overallWeight))
			selectionValue, err := cryptoRand.Int(cryptoRand.Reader, maxBound)
			if err != nil {
				return newError("failed to generate random selection value").Base(err).AtWarning()
			}
			selectedValue := int32(selectionValue.Int64())
			currentValue := int32(0)
			for _, nextStep := range step.NextStep {
				if currentValue >= selectedValue {
					currentStep = int(nextStep.GotoLocation)
					break
				}
				currentValue += nextStep.Weight
			}
			newError("invalid steps instruction, check configuration for step", currentStep).AtError().WriteToLog()
			currentStep++
		}
	}
}

func (generator *TrafficGenerator) tlsHandshake(conn net.Conn) (security.Conn, error) {
	securityEngine, err := common.CreateObject(generator.ctx, generator.config.SecuritySettings)
	if err != nil {
		return nil, newError("unable to create security engine from security settings").Base(err)
	}
	securityEngineTyped, ok := securityEngine.(security.Engine)
	if !ok {
		return nil, newError("type assertion error when create security engine from security settings")
	}

	return securityEngineTyped.Client(conn)
}

type httpRequestTransport interface {
	http.RoundTripper
}

func newHTTPRequestTransportH1(conn net.Conn) httpRequestTransport {
	return &httpRequestTransportH1{
		conn:      conn,
		bufReader: bufio.NewReader(conn),
	}
}

type httpRequestTransportH1 struct {
	conn      net.Conn
	bufReader *bufio.Reader
}

func (h *httpRequestTransportH1) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1

	err := req.Write(h.conn)
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(h.bufReader, req)
}

func newHTTPRequestTransportH2(conn net.Conn) httpRequestTransport {
	transport := &http2.Transport{}
	clientConn, err := transport.NewClientConn(conn)
	if err != nil {
		return nil
	}
	return &httpRequestTransportH2{
		transport:        transport,
		clientConnection: clientConn,
	}
}

type httpRequestTransportH2 struct {
	transport        *http2.Transport
	clientConnection *http2.ClientConn
}

func (h *httpRequestTransportH2) RoundTrip(request *http.Request) (*http.Response, error) {
	request.ProtoMajor = 2
	request.ProtoMinor = 0

	response, err := h.clientConnection.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func newSingleConnectionHTTPTransport(conn net.Conn, alpn string) (httpRequestTransport, error) {
	switch alpn {
	case "h2":
		return newHTTPRequestTransportH2(conn), nil
	case "http/1.1", "":
		return newHTTPRequestTransportH1(conn), nil
	default:
		return nil, newError("unknown alpn: " + alpn).AtWarning()
	}
}
