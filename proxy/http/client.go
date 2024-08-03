package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/http2"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/bytespool"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/retry"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/proxy"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
)

type Client struct {
	serverPicker       protocol.ServerPicker
	policyManager      policy.Manager
	h1SkipWaitForReply bool
}

type h2Conn struct {
	rawConn net.Conn
	h2Conn  *http2.ClientConn
}

var (
	cachedH2Mutex sync.Mutex
	cachedH2Conns map[net.Destination]h2Conn
)

// NewClient create a new http client based on the given config.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	serverList := protocol.NewServerList()
	for _, rec := range config.Server {
		s, err := protocol.NewServerSpecFromPB(rec)
		if err != nil {
			return nil, newError("failed to get server spec").Base(err)
		}
		serverList.AddServer(s)
	}
	if serverList.Size() == 0 {
		return nil, newError("0 target server")
	}

	v := core.MustFromContext(ctx)
	return &Client{
		serverPicker:       protocol.NewRoundRobinServerPicker(serverList),
		policyManager:      v.GetFeature(policy.ManagerType()).(policy.Manager),
		h1SkipWaitForReply: config.H1SkipWaitForReply,
	}, nil
}

// Process implements proxy.Outbound.Process. We first create a socket tunnel via HTTP CONNECT method, then redirect all inbound traffic to that tunnel.
func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified.")
	}
	target := outbound.Target
	targetAddr := target.NetAddr()

	if target.Network == net.Network_UDP {
		return newError("UDP is not supported by HTTP outbound")
	}

	var user *protocol.MemoryUser
	var conn internet.Connection

	var firstPayload []byte

	if reader, ok := link.Reader.(buf.TimeoutReader); ok {
		// 0-RTT optimization for HTTP/2: If the payload comes very soon, it can be
		// transmitted together. Note we should not get stuck here, as the payload may
		// not exist (considering to access MySQL database via a HTTP proxy, where the
		// server sends hello to the client first).
		waitTime := proxy.FirstPayloadTimeout
		if c.h1SkipWaitForReply {
			// Some server require first write to be present in client hello.
			// Increase timeout to if the client have explicitly requested to skip waiting for reply.
			waitTime = time.Second
		}
		if mbuf, _ := reader.ReadMultiBufferTimeout(waitTime); mbuf != nil {
			mlen := mbuf.Len()
			firstPayload = bytespool.Alloc(mlen)
			mbuf, _ = buf.SplitBytes(mbuf, firstPayload)
			firstPayload = firstPayload[:mlen]

			buf.ReleaseMulti(mbuf)
			defer bytespool.Free(firstPayload)
		}
	}

	if err := retry.ExponentialBackoff(5, 100).On(func() error {
		server := c.serverPicker.PickServer()
		dest := server.Destination()
		user = server.PickUser()

		netConn, firstResp, err := setUpHTTPTunnel(ctx, dest, targetAddr, user, dialer, firstPayload, c.h1SkipWaitForReply)
		if netConn != nil {
			if _, ok := netConn.(*http2Conn); !ok && !c.h1SkipWaitForReply {
				if _, err := netConn.Write(firstPayload); err != nil {
					netConn.Close()
					return err
				}
			}
			if firstResp != nil {
				if err := link.Writer.WriteMultiBuffer(firstResp); err != nil {
					return err
				}
			}
			conn = internet.Connection(netConn)
		}
		return err
	}); err != nil {
		return newError("failed to find an available destination").Base(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			newError("failed to closed connection").Base(err).WriteToLog(session.ExportIDToError(ctx))
		}
	}()

	p := c.policyManager.ForLevel(0)
	if user != nil {
		p = c.policyManager.ForLevel(user.Level)
	}

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, p.Timeouts.ConnectionIdle)

	requestFunc := func() error {
		defer timer.SetTimeout(p.Timeouts.DownlinkOnly)
		return buf.Copy(link.Reader, buf.NewWriter(conn), buf.UpdateActivity(timer))
	}
	responseFunc := func() error {
		defer timer.SetTimeout(p.Timeouts.UplinkOnly)
		return buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer))
	}

	responseDonePost := task.OnSuccess(responseFunc, task.Close(link.Writer))
	if err := task.Run(ctx, requestFunc, responseDonePost); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

// setUpHTTPTunnel will create a socket tunnel via HTTP CONNECT method
func setUpHTTPTunnel(ctx context.Context, dest net.Destination, target string, user *protocol.MemoryUser, dialer internet.Dialer, firstPayload []byte, writeFirstPayloadInH1 bool,
) (net.Conn, buf.MultiBuffer, error) {
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: target},
		Header: make(http.Header),
		Host:   target,
	}

	if user != nil && user.Account != nil {
		account := user.Account.(*Account)
		auth := account.GetUsername() + ":" + account.GetPassword()
		req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}

	connectHTTP1 := func(rawConn net.Conn) (net.Conn, buf.MultiBuffer, error) {
		req.Header.Set("Proxy-Connection", "Keep-Alive")

		if !writeFirstPayloadInH1 {
			err := req.Write(rawConn)
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
		} else {
			buffer := bytes.NewBuffer(nil)
			err := req.Write(buffer)
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
			_, err = io.Copy(buffer, bytes.NewReader(firstPayload))
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
			_, err = rawConn.Write(buffer.Bytes())
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
		}
		bufferedReader := bufio.NewReader(rawConn)
		resp, err := http.ReadResponse(bufferedReader, req)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			rawConn.Close()
			return nil, nil, newError("Proxy responded with non 200 code: " + resp.Status)
		}
		if bufferedReader.Buffered() > 0 {
			payload, err := buf.ReadFrom(io.LimitReader(bufferedReader, int64(bufferedReader.Buffered())))
			if err != nil {
				return nil, nil, newError("unable to drain buffer: ").Base(err)
			}
			return rawConn, payload, nil
		}
		return rawConn, nil, nil
	}

	connectHTTP2 := func(rawConn net.Conn, h2clientConn *http2.ClientConn) (net.Conn, error) {
		pr, pw := io.Pipe()
		req.Body = pr

		var pErr error
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			_, pErr = pw.Write(firstPayload)
			wg.Done()
		}()

		resp, err := h2clientConn.RoundTrip(req) // nolint: bodyclose
		if err != nil {
			rawConn.Close()
			return nil, err
		}

		wg.Wait()
		if pErr != nil {
			rawConn.Close()
			return nil, pErr
		}

		if resp.StatusCode != http.StatusOK {
			rawConn.Close()
			return nil, newError("Proxy responded with non 200 code: " + resp.Status)
		}
		return newHTTP2Conn(rawConn, pw, resp.Body), nil
	}

	cachedH2Mutex.Lock()
	cachedConn, cachedConnFound := cachedH2Conns[dest]
	cachedH2Mutex.Unlock()

	if cachedConnFound {
		rc, cc := cachedConn.rawConn, cachedConn.h2Conn
		if cc.CanTakeNewRequest() {
			proxyConn, err := connectHTTP2(rc, cc)
			if err != nil {
				return nil, nil, err
			}

			return proxyConn, nil, nil
		}
	}

	rawConn, err := dialer.Dial(ctx, dest)
	if err != nil {
		return nil, nil, err
	}

	iConn := rawConn
	if statConn, ok := iConn.(*internet.StatCouterConnection); ok {
		iConn = statConn.Connection
	}

	nextProto := ""
	if connALPNGetter, ok := iConn.(security.ConnectionApplicationProtocol); ok {
		nextProto, err = connALPNGetter.GetConnectionApplicationProtocol()
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}
	}

	switch nextProto {
	case "", "http/1.1":
		return connectHTTP1(rawConn)
	case "h2":
		t := http2.Transport{}
		h2clientConn, err := t.NewClientConn(rawConn)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}

		proxyConn, err := connectHTTP2(rawConn, h2clientConn)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}

		cachedH2Mutex.Lock()
		if cachedH2Conns == nil {
			cachedH2Conns = make(map[net.Destination]h2Conn)
		}

		cachedH2Conns[dest] = h2Conn{
			rawConn: rawConn,
			h2Conn:  h2clientConn,
		}
		cachedH2Mutex.Unlock()

		return proxyConn, nil, err
	default:
		return nil, nil, newError("negotiated unsupported application layer protocol: " + nextProto)
	}
}

func newHTTP2Conn(c net.Conn, pipedReqBody *io.PipeWriter, respBody io.ReadCloser) net.Conn {
	return &http2Conn{Conn: c, in: pipedReqBody, out: respBody}
}

type http2Conn struct {
	net.Conn
	in  *io.PipeWriter
	out io.ReadCloser
}

func (h *http2Conn) Read(p []byte) (n int, err error) {
	return h.out.Read(p)
}

func (h *http2Conn) Write(p []byte) (n int, err error) {
	return h.in.Write(p)
}

func (h *http2Conn) Close() error {
	h.in.Close()
	return h.out.Close()
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
