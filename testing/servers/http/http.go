package tcp

import (
	"net/http"

	"github.com/v2fly/v2ray-core/v5/common/net"
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
	if !found {
		// Without this the response defaults to an empty 200, which hides
		// requests sent to a path the test did not register a handler for.
		resp.WriteHeader(http.StatusNotFound)
		return
	}
	handler(resp, req)
}

func (s *Server) Start() (net.Destination, error) {
	// The listener is created before returning, otherwise a caller may connect
	// before the socket exists, and a failure to bind would go unnoticed.
	listener, err := net.Listen("tcp", "127.0.0.1:"+s.Port.String())
	if err != nil {
		return net.Destination{}, err
	}

	localAddr := listener.Addr().(*net.TCPAddr)
	s.Port = net.Port(localAddr.Port)
	s.server = &http.Server{Handler: s}
	go s.server.Serve(listener)

	return net.TCPDestination(net.LocalHostIP, s.Port), nil
}

func (s *Server) Close() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
