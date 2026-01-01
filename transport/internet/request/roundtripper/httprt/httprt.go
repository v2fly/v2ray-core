package httprt

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	gonet "net"
	"net/http"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func newHTTPRoundTripperClient(ctx context.Context, config *ClientConfig) request.RoundTripperClient {
	_ = ctx
	return &httpTripperClient{config: config}
}

type httpTripperClient struct {
	httpRTT  http.RoundTripper
	config   *ClientConfig
	assembly request.TransportClientAssembly
}

type unimplementedBackDrop struct{}

func (u unimplementedBackDrop) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, newError("unimplemented")
}

func (h *httpTripperClient) OnTransportClientAssemblyReady(assembly request.TransportClientAssembly) {
	h.assembly = assembly
}

func (h *httpTripperClient) RoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	var streamingWriter io.Writer
	for _, v := range opts {
		if streamingResp, ok := v.(request.OptionSupportsStreamingResponse); ok {
			streamingWriter = streamingResp.GetResponseWriter()
		}
	}
	if h.httpRTT == nil {
		var backDrop http.RoundTripper = unimplementedBackDrop{}
		if h.config.AllowHttp {
			backDrop = &http.Transport{
				DialContext: func(_ context.Context, network, addr string) (gonet.Conn, error) {
					return h.assembly.AutoImplDialer().Dial(ctx)
				},
				DialTLSContext: func(_ context.Context, network, addr string) (gonet.Conn, error) {
					return nil, newError("unexpected dial of TLS")
				},
			}
		}
		h.httpRTT = transportcommon.NewALPNAwareHTTPRoundTripperWithH2Pool(ctx, func(ctx context.Context, addr string) (gonet.Conn, error) {
			return h.assembly.AutoImplDialer().Dial(ctx)
		}, backDrop, int(h.config.H2PoolSize))
	}

	connectionTagStr := base64.RawURLEncoding.EncodeToString(req.ConnectionTag)

	httpRequest, err := http.NewRequest("POST", h.config.Http.UrlPrefix+h.config.Http.Path, bytes.NewReader(req.Data))
	if err != nil {
		return resp, err
	}

	httpRequest.Header.Set("X-Session-ID", connectionTagStr)

	httpResp, err := h.httpRTT.RoundTrip(httpRequest)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		newError("non-200 response: ", httpResp.Status).AtInfo().WriteToLog()
	}
	if streamingWriter == nil {
		result, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return request.Response{}, err
		}
		return request.Response{Data: result}, nil
	}
	_, err = io.Copy(streamingWriter, httpResp.Body)
	if err != nil {
		return request.Response{}, newError("unable to copy response").Base(err)
	}
	return request.Response{}, nil
}

func newHTTPRoundTripperServer(ctx context.Context, config *ServerConfig) request.RoundTripperServer {
	return &httpTripperServer{ctx: ctx, config: config}
}

type httpTripperServer struct {
	ctx      context.Context
	listener net.Listener
	assembly request.TransportServerAssembly

	config *ServerConfig
}

func (h *httpTripperServer) OnTransportServerAssemblyReady(assembly request.TransportServerAssembly) {
	h.assembly = assembly
}

func (h *httpTripperServer) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	h.onRequest(writer, r)
}

type httpRespStreamWriting struct {
	resp http.ResponseWriter
	used bool
}

func (h *httpRespStreamWriting) RoundTripperOption() {
}

func (h *httpRespStreamWriting) GetResponseWriter() io.Writer {
	h.used = true
	return h.resp
}

func (h *httpRespStreamWriting) Flush() {
	if f, ok := h.resp.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *httpTripperServer) onRequest(resp http.ResponseWriter, req *http.Request) {
	tail := req.Header.Get("X-Session-ID")
	data := []byte(tail)
	if !h.config.NoDecodingSessionTag {
		decodedData, err := base64.RawURLEncoding.DecodeString(tail)
		if err != nil {
			newError("unable to decode tag").Base(err).AtInfo().WriteToLog()
			return
		}
		data = decodedData
	}
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		newError("unable to read body").Base(err).AtInfo().WriteToLog()
	}

	streamingRespOption := &httpRespStreamWriting{resp: resp}
	recvResp, err := h.assembly.TripperReceiver().OnRoundTrip(h.ctx, request.Request{Data: body, ConnectionTag: data},
		streamingRespOption)
	if err != nil {
		newError("unable to process roundtrip").Base(err).AtInfo().WriteToLog()
	}
	if streamingRespOption.used {
		return
	}
	_, err = io.Copy(resp, bytes.NewReader(recvResp.Data))
	if err != nil {
		newError("unable to send response").Base(err).AtInfo().WriteToLog()
	}
}

func (h *httpTripperServer) Start() error {
	listener, err := h.assembly.AutoImplListener().Listen(h.ctx)
	if err != nil {
		return newError("unable to create a listener for http tripper server").Base(err)
	}
	h.listener = listener
	go func() {
		httpServer := http.Server{
			ReadHeaderTimeout: 240 * time.Second,
			ReadTimeout:       240 * time.Second,
			WriteTimeout:      240 * time.Second,
			IdleTimeout:       240 * time.Second,
		}
		httpServer.Handler = h
		err := httpServer.Serve(h.listener)
		if err != nil {
			newError("unable to serve listener for http tripper server").Base(err).WriteToLog()
		}
	}()
	return nil
}

func (h *httpTripperServer) Close() error {
	return h.listener.Close()
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		clientConfig, ok := config.(*ClientConfig)
		if !ok {
			return nil, newError("not a ClientConfig")
		}
		return newHTTPRoundTripperClient(ctx, clientConfig), nil
	}))
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		serverConfig, ok := config.(*ServerConfig)
		if !ok {
			return nil, newError("not a ServerConfig")
		}
		return newHTTPRoundTripperServer(ctx, serverConfig), nil
	}))
}
