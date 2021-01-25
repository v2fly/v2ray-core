package router

import (
	"context"
	"math"
	"net/http"
	"time"

	"v2ray.com/core/common/net"
	"v2ray.com/core/common/session"
	"v2ray.com/core/features/routing"
)

type pingClient struct {
	httpClient *http.Client

	Dispatcher  routing.Dispatcher
	Handler     string
	Destination string
	Timeout     time.Duration
}

func (s *pingClient) newHTTPClient() *http.Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := net.ParseDestination(network + ":" + addr)
			if err != nil {
				return nil, err
			}
			h := &session.Handler{
				Tag: s.Handler,
			}
			ctx = session.ContextWithHandler(ctx, h)
			link, err := s.Dispatcher.Dispatch(ctx, dest)
			if err != nil {
				return nil, err
			}
			return net.NewConnection(
				net.ConnectionInputMulti(link.Writer),
				net.ConnectionOutputMulti(link.Reader),
			), nil
		},
	}
	return &http.Client{
		Transport: tr,
		Timeout:   s.Timeout,
		// don't follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// MeasureDelay returns the delay time of the request to dest
func (s *pingClient) MeasureDelay() (time.Duration, error) {
	if s.httpClient == nil {
		s.httpClient = s.newHTTPClient()
	}
	req, err := http.NewRequest(http.MethodHead, s.Destination, nil)
	if err != nil {
		return math.MaxInt64, err
	}
	start := time.Now()
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return math.MaxInt64, err
	}
	// don't wait for body
	resp.Body.Close()
	return time.Since(start), nil
}
