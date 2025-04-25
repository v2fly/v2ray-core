package tcp

import (
	"net/http"

	"github.com/ghxhy/v2ray-core/v5/common/net"
)

type Server struct {
	Port        net.Port
	PathHandler map[string]http.HandlerFunc
	server      *http.Server
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("Home"))
		return
	}

	handler, found := s.PathHandler[req.URL.Path]
	if found {
		handler(resp, req)
	}
}

func (s *Server) Start() (net.Destination, error) {
	s.server = &http.Server{
		Addr:    "127.0.0.1:" + s.Port.String(),
		Handler: s,
	}
	go s.server.ListenAndServe()
	return net.TCPDestination(net.LocalHostIP, s.Port), nil
}

func (s *Server) Close() error {
	return s.server.Close()
}
