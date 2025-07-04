package httponconnection

import (
	"bufio"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type HttpRequestTransport interface {
	http.RoundTripper
}

func newHTTPRequestTransportH1(conn net.Conn) HttpRequestTransport {
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

func newHTTPRequestTransportH2(conn net.Conn) HttpRequestTransport {
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

func newSingleConnectionHTTPTransport(conn net.Conn, alpn string) (HttpRequestTransport, error) {
	switch alpn {
	case "h2":
		return newHTTPRequestTransportH2(conn), nil
	case "http/1.1", "":
		return newHTTPRequestTransportH1(conn), nil
	default:
		return nil, newError("unknown alpn: " + alpn).AtWarning()
	}
}

func NewSingleConnectionHTTPTransport(conn net.Conn, alpn string) (HttpRequestTransport, error) {
	return newSingleConnectionHTTPTransport(conn, alpn)
}
