package httpenrollmentconfirmation

import (
	"encoding/base32"
	"net"
	"net/http"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/httponconnection"
)

func NewClientRoundTripperForEnrollmentConfirmation(
	dial func(network, addr string) (net.Conn, error),
	serverIdentity []byte,
) (http.RoundTripper, RoundTripperMetadata, error) {
	if dial == nil {
		return nil, nil, newError("nil dial function")
	}
	if len(serverIdentity) == 0 {
		return nil, nil, newError("nil or empty server identity")
	}
	cr := &clientRoundtripper{
		dial:           dial,
		serverIdentity: serverIdentity,
	}
	return cr, cr, nil
}

type RoundTripperMetadata interface {
	IsCreatingSecondaryNewConnection() bool
}

type clientRoundtripper struct {
	dial           func(network, addr string) (net.Conn, error)
	serverIdentity []byte

	currentConnInnerConn common.Closable
	currentConn          http.RoundTripper
	currentConnLock      sync.RWMutex

	pendingNewConnection int
	// DO NOT ATTEMPT TO ACQUIRE ANY LOCK WHILE HOLDING THIS LOCK
	pendingNewConnectionLock sync.RWMutex
}

func (c *clientRoundtripper) IsCreatingSecondaryNewConnection() bool {
	defer c.pendingNewConnectionLock.RUnlock()
	c.pendingNewConnectionLock.RLock()
	return c.pendingNewConnection >= 1
}

func (c *clientRoundtripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if c.IsCreatingSecondaryNewConnection() {
		return nil, newError("another connection is being established, cannot create a secondary connection")
	}
	return c.roundTrip(request)
}

func (c *clientRoundtripper) roundTrip(request *http.Request) (*http.Response, error) {
	c.currentConnLock.RLock()

	if c.currentConn == nil {
		c.currentConnLock.RUnlock()
		c.pendingNewConnectionLock.Lock()
		c.pendingNewConnection += 1
		c.pendingNewConnectionLock.Unlock()
		decreaseCount := func() {
			c.pendingNewConnectionLock.Lock()
			c.pendingNewConnection -= 1
			c.pendingNewConnectionLock.Unlock()
		}
		if err := c.createNewConnection(); err != nil {
			decreaseCount()
			return nil, err // Failed to create a new connection
		}
		resp, err := c.roundTrip(request)
		decreaseCount()
		return resp, err
	}
	defer c.currentConnLock.RUnlock()

	// Use the current connection to perform the round trip
	resp, err := c.currentConn.RoundTrip(request)
	if err != nil {
		defer func() {
			c.currentConnLock.RUnlock()
			c.currentConnLock.Lock()
			if c.currentConn != nil {
				c.currentConnInnerConn.Close()
				c.currentConnInnerConn = nil
				c.currentConn = nil
			}
			c.currentConnLock.Unlock()
			c.currentConnLock.RLock()
		}()
		return nil, newError("unable to roundtrip for enrollment verification").Base(err)
	}
	return resp, nil
}

func (c *clientRoundtripper) createNewConnection() error {
	c.currentConnLock.Lock()
	defer c.currentConnLock.Unlock()

	if c.currentConn != nil {
		return nil // Connection already exists
	}

	serverID := base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding).EncodeToString(c.serverIdentity)
	conn, err := c.dial("tcp", serverID+tlsmirror.EnrollmentVerificationControlConnectionPostfix+":80")
	if err != nil {
		return newError("failed to dial server: ", err)
	}
	c.currentConnInnerConn = conn
	c.currentConn, err = httponconnection.NewSingleConnectionHTTPTransport(conn, "h2")
	if err != nil {
		conn.Close() // Close the connection if transport creation fails
		return newError("failed to create HTTP transport: ", err)
	}
	return nil
}
