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
) (http.RoundTripper, error) {
	if dial == nil {
		return nil, newError("nil dial function")
	}
	if len(serverIdentity) == 0 {
		return nil, newError("nil or empty server identity")
	}
	return &clientRoundtripper{
		dial:           dial,
		serverIdentity: serverIdentity,
	}, nil
}

type clientRoundtripper struct {
	dial           func(network, addr string) (net.Conn, error)
	serverIdentity []byte

	currentConnInnerConn common.Closable
	currentConn          http.RoundTripper
	currentConnLock      sync.RWMutex
}

func (c *clientRoundtripper) RoundTrip(request *http.Request) (*http.Response, error) {
	c.currentConnLock.RLock()

	if c.currentConn == nil {
		c.currentConnLock.RUnlock()
		if err := c.createNewConnection(); err != nil {
			return nil, err // Failed to create a new connection
		}
		return c.RoundTrip(request)
	}
	defer c.currentConnLock.RUnlock()

	// Use the current connection to perform the round trip
	resp, err := c.currentConn.RoundTrip(request)
	if err != nil {
		defer func() {
			c.currentConnLock.Lock()
			defer c.currentConnLock.Unlock()
			if c.currentConn != nil {
				c.currentConnInnerConn.Close()
				c.currentConnInnerConn = nil
				c.currentConn = nil
			}
		}()
		return nil, newError("unable to roundtrip for enrollment verification").Base(err)
	}
	return resp, err
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
