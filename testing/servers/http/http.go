package tcp

import (
	"net/http"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type Server struct {
	Port        net.Port
	PathHandler map[string]http.HandlerFunc
	server      *http.Server
	listener    net.Listener
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
	listener, err := net.Listen("tcp", "127.0.0.1:"+s.Port.String())
	if err != nil {
		return net.Destination{}, err
	}
	s.Port = net.Port(listener.Addr().(*net.TCPAddr).Port)
	s.listener = listener
	s.server = &http.Server{
		Handler: s,
	}
	go s.server.Serve(listener)
	return net.TCPDestination(net.LocalHostIP, s.Port), nil
}

func (s *Server) Close() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
