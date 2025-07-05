package httpenrollmentconfirmation

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func NewHTTPConnectionHub(handler http.Handler) *HTTPConnectionHub {
	return &HTTPConnectionHub{
		handler:  handler,
		h2server: &http2.Server{},
	}
}

type HTTPConnectionHub struct {
	handler  http.Handler
	h2server *http2.Server
}

func (h *HTTPConnectionHub) ServeConnection(ctx context.Context, conn net.Conn) error {
	go h.h2server.ServeConn(conn, &http2.ServeConnOpts{
		Handler: h.handler,
	})
	return nil
}
