package transportcommon

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"

	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
)

type DialerFunc func(ctx context.Context, addr string) (net.Conn, error)

// NewALPNAwareHTTPRoundTripper creates an instance of RoundTripper that dial to remote HTTPS endpoint with
// an alternative version of TLS implementation.
func NewALPNAwareHTTPRoundTripper(ctx context.Context, dialer DialerFunc,
	backdropTransport http.RoundTripper,
) http.RoundTripper {
	rtImpl := &alpnAwareHTTPRoundTripperImpl{
		connectWithH1:     map[string]bool{},
		backdropTransport: backdropTransport,
		pendingConn:       map[pendingConnKey]*unclaimedConnection{},
		dialer:            dialer,
		ctx:               ctx,
	}
	rtImpl.init()
	return rtImpl
}

type alpnAwareHTTPRoundTripperImpl struct {
	accessConnectWithH1 sync.Mutex
	connectWithH1       map[string]bool

	httpsH1Transport  http.RoundTripper
	httpsH2Transport  http.RoundTripper
	backdropTransport http.RoundTripper

	accessDialingConnection sync.Mutex
	pendingConn             map[pendingConnKey]*unclaimedConnection

	ctx    context.Context
	dialer DialerFunc
}

type pendingConnKey struct {
	isH2 bool
	dest string
}

var (
	errEAGAIN        = errors.New("incorrect ALPN negotiated, try again with another ALPN")
	errEAGAINTooMany = errors.New("incorrect ALPN negotiated")
	errExpired       = errors.New("connection have expired")
)

func (r *alpnAwareHTTPRoundTripperImpl) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return r.backdropTransport.RoundTrip(req)
	}
	for retryCount := 0; retryCount < 5; retryCount++ {
		effectivePort := req.URL.Port()
		if effectivePort == "" {
			effectivePort = "443"
		}
		if r.getShouldConnectWithH1(fmt.Sprintf("%v:%v", req.URL.Hostname(), effectivePort)) {
			resp, err := r.httpsH1Transport.RoundTrip(req)
			if errors.Is(err, errEAGAIN) {
				continue
			}
			return resp, err
		}
		resp, err := r.httpsH2Transport.RoundTrip(req)
		if errors.Is(err, errEAGAIN) {
			continue
		}
		return resp, err
	}
	return nil, errEAGAINTooMany
}

func (r *alpnAwareHTTPRoundTripperImpl) getShouldConnectWithH1(domainName string) bool {
	r.accessConnectWithH1.Lock()
	defer r.accessConnectWithH1.Unlock()
	if value, set := r.connectWithH1[domainName]; set {
		return value
	}
	return false
}

func (r *alpnAwareHTTPRoundTripperImpl) setShouldConnectWithH1(domainName string) {
	r.accessConnectWithH1.Lock()
	defer r.accessConnectWithH1.Unlock()
	r.connectWithH1[domainName] = true
}

func (r *alpnAwareHTTPRoundTripperImpl) clearShouldConnectWithH1(domainName string) {
	r.accessConnectWithH1.Lock()
	defer r.accessConnectWithH1.Unlock()
	r.connectWithH1[domainName] = false
}

func getPendingConnectionID(dest string, alpnIsH2 bool) pendingConnKey {
	return pendingConnKey{isH2: alpnIsH2, dest: dest}
}

func (r *alpnAwareHTTPRoundTripperImpl) putConn(addr string, alpnIsH2 bool, conn net.Conn) {
	connID := getPendingConnectionID(addr, alpnIsH2)
	r.pendingConn[connID] = NewUnclaimedConnection(conn, time.Minute)
}

func (r *alpnAwareHTTPRoundTripperImpl) getConn(addr string, alpnIsH2 bool) net.Conn {
	connID := getPendingConnectionID(addr, alpnIsH2)
	if conn, ok := r.pendingConn[connID]; ok {
		delete(r.pendingConn, connID)
		if claimedConnection, err := conn.claimConnection(); err == nil {
			return claimedConnection
		}
	}
	return nil
}

func (r *alpnAwareHTTPRoundTripperImpl) dialOrGetTLSWithExpectedALPN(ctx context.Context, addr string, expectedH2 bool) (net.Conn, error) {
	r.accessDialingConnection.Lock()
	defer r.accessDialingConnection.Unlock()

	if r.getShouldConnectWithH1(addr) == expectedH2 {
		return nil, errEAGAIN
	}

	// Get a cached connection if possible to reduce preflight connection closed without sending data
	if gconn := r.getConn(addr, expectedH2); gconn != nil {
		return gconn, nil
	}

	conn, err := r.dialTLS(ctx, addr)
	if err != nil {
		return nil, err
	}

	protocol := ""
	if connAPLNGetter, ok := conn.(security.ConnectionApplicationProtocol); ok {
		connectionALPN, err := connAPLNGetter.GetConnectionApplicationProtocol()
		if err != nil {
			return nil, newError("failed to get connection ALPN").Base(err).AtWarning()
		}
		protocol = connectionALPN
	}

	protocolIsH2 := protocol == http2.NextProtoTLS

	if protocolIsH2 == expectedH2 {
		return conn, err
	}

	r.putConn(addr, protocolIsH2, conn)

	if protocolIsH2 {
		r.clearShouldConnectWithH1(addr)
	} else {
		r.setShouldConnectWithH1(addr)
	}

	return nil, errEAGAIN
}

func (r *alpnAwareHTTPRoundTripperImpl) dialTLS(ctx context.Context, addr string) (net.Conn, error) {
	_ = ctx
	return r.dialer(r.ctx, addr)
}

func (r *alpnAwareHTTPRoundTripperImpl) init() {
	r.httpsH2Transport = &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return r.dialOrGetTLSWithExpectedALPN(context.Background(), addr, true)
		},
	}
	r.httpsH1Transport = &http.Transport{
		DialTLSContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return r.dialOrGetTLSWithExpectedALPN(ctx, addr, false)
		},
	}
}

func NewUnclaimedConnection(conn net.Conn, expireTime time.Duration) *unclaimedConnection {
	c := &unclaimedConnection{
		Conn: conn,
	}
	time.AfterFunc(expireTime, c.tick)
	return c
}

type unclaimedConnection struct {
	net.Conn
	claimed bool
	access  sync.Mutex
}

func (c *unclaimedConnection) claimConnection() (net.Conn, error) {
	c.access.Lock()
	defer c.access.Unlock()
	if !c.claimed {
		c.claimed = true
		return c.Conn, nil
	}
	return nil, errExpired
}

func (c *unclaimedConnection) tick() {
	c.access.Lock()
	defer c.access.Unlock()
	if !c.claimed {
		c.claimed = true
		c.Conn.Close()
		c.Conn = nil
	}
}
