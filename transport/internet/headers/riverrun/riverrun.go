package riverrun

import (
	"context"
	"fmt"
	gonet "net"
	"time"

	"github.com/v2fly/riverrun"
	"github.com/v2fly/riverrun/common/drbg"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type errConn struct {
	err error
}

func (e errConn) Read(b []byte) (n int, err error) {
	return 0, e.err
}

func (e errConn) Write(b []byte) (n int, err error) {
	return 0, e.err
}

func (e errConn) Close() error {
	return e.err
}

func (e errConn) LocalAddr() gonet.Addr {
	return &gonet.UnixAddr{Name: "error"}
}

func (e errConn) RemoteAddr() gonet.Addr {
	return &gonet.UnixAddr{Name: "error"}
}

func (e errConn) SetDeadline(t time.Time) error {
	return e.err
}

func (e errConn) SetReadDeadline(t time.Time) error {
	return e.err
}

func (e errConn) SetWriteDeadline(t time.Time) error {
	return e.err
}

type riverrunConnectionFactory struct {
	config *Config
}

func (p riverrunConnectionFactory) Infof(format string, a ...interface{}) {
	newError(fmt.Sprintf(format, a...)).AtInfo().WriteToLog()
}

func (p riverrunConnectionFactory) Debugf(format string, a ...interface{}) {
	newError(fmt.Sprintf(format, a...)).AtDebug().WriteToLog()
}

func (p riverrunConnectionFactory) Client(conn net.Conn) net.Conn {
	seed, err := drbg.SeedFromBytes([]byte(p.config.Seed))
	if err != nil {
		return errConn{err: err}
	}
	wconn, err := riverrun.NewConn(conn, false, seed, p)
	if err != nil {
		return errConn{err: err}
	}
	return wconn
}

func (p riverrunConnectionFactory) Server(conn net.Conn) net.Conn {
	seed, err := drbg.SeedFromBytes([]byte(p.config.Seed))
	if err != nil {
		return errConn{err: err}
	}
	wconn, err := riverrun.NewConn(conn, true, seed, p)
	if err != nil {
		return errConn{err: err}
	}
	return wconn
}

func newRiverrunConnectionAuthenticator(config *Config) (internet.ConnectionAuthenticator, error) {
	return riverrunConnectionFactory{
		config: config,
	}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return newRiverrunConnectionAuthenticator(config.(*Config))
	}))
}
