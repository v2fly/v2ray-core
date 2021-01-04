package router

import (
	"context"
	"fmt"
	"io/ioutil"
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
	}
}

func (s *pingClient) doRequest() (int, []byte, error) {
	if s.httpClient == nil {
		s.httpClient = s.newHTTPClient()
	}
	req, err := http.NewRequest("GET", s.Destination, nil)
	if err != nil {
		return -1, nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

// MeasureDelay returns the delay time of the request to dest
func (s *pingClient) MeasureDelay() (time.Duration, error) {
	start := time.Now()
	code, _, err := s.doRequest()
	if err != nil {
		return -1, err
	}
	if code > 399 {
		return -1, fmt.Errorf("status incorrect (>= 400): %d", code)
	}
	return time.Since(start), nil
}
