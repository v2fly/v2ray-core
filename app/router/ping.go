package router

import (
	"context"
	"net/http"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/features/routing"
)

type pingClient struct {
	destination string
	httpClient  *http.Client
}

func newPingClient(destination string, timeout time.Duration, handler string, dispatcher routing.Dispatcher) *pingClient {
	return &pingClient{
		destination: destination,
		httpClient:  newHTTPClient(handler, dispatcher, timeout),
	}
}

func newDirectPingClient(destination string, timeout time.Duration) *pingClient {
	return &pingClient{
		destination: destination,
		httpClient:  &http.Client{Timeout: timeout},
	}
}

func newHTTPClient(handler string, dispatcher routing.Dispatcher, timeout time.Duration) *http.Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := net.ParseDestination(network + ":" + addr)
			if err != nil {
				return nil, err
			}
			h := &session.Handler{
				Tag: handler,
			}
			ctx = session.ContextWithHandler(ctx, h)
			link, err := dispatcher.Dispatch(ctx, dest)
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
		Timeout:   timeout,
		// don't follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// MeasureDelay returns the delay time of the request to dest
func (s *pingClient) MeasureDelay() (time.Duration, error) {
	if s.httpClient == nil {
		panic("pingClient no initialized")
	}
	req, err := http.NewRequest(http.MethodHead, s.destination, nil)
	if err != nil {
		return rttFailed, err
	}
	start := time.Now()
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return rttFailed, err
	}
	// don't wait for body
	resp.Body.Close()
	return time.Since(start), nil
}
