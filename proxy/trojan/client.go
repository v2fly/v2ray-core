// +build !confonly

package trojan

import (
	"context"
	"syscall"
	"time"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/platform"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/retry"
	"v2ray.com/core/common/session"
	"v2ray.com/core/common/signal"
	"v2ray.com/core/common/task"
	"v2ray.com/core/features/policy"
	"v2ray.com/core/transport"
	"v2ray.com/core/transport/internet"
	"v2ray.com/core/transport/internet/xtls"
)

// Client is a inbound handler for trojan protocol
type Client struct {
	serverPicker  protocol.ServerPicker
	policyManager policy.Manager
}

// NewClient create a new trojan client.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	serverList := protocol.NewServerList()
	for _, rec := range config.Server {
		s, err := protocol.NewServerSpecFromPB(rec)
		if err != nil {
			return nil, newError("failed to parse server spec").Base(err)
		}
		serverList.AddServer(s)
	}
	if serverList.Size() == 0 {
		return nil, newError("0 server")
	}

	v := core.MustFromContext(ctx)
	client := &Client{
		serverPicker:  protocol.NewRoundRobinServerPicker(serverList),
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}
	return client, nil
}

// Process implements OutboundHandler.Process().
func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified")
	}
	destination := outbound.Target
	network := destination.Network

	var server *protocol.ServerSpec
	var conn internet.Connection

	err := retry.ExponentialBackoff(5, 100).On(func() error {
		server = c.serverPicker.PickServer()
		rawConn, err := dialer.Dial(ctx, server.Destination())
		if err != nil {
			return err
		}

		conn = rawConn
		return nil
	})
	if err != nil {
		return newError("failed to find an available destination").AtWarning().Base(err)
	}
	newError("tunneling request to ", destination, " via ", server.Destination()).WriteToLog(session.ExportIDToError(ctx))

	defer conn.Close()

	user := server.PickUser()
	account, ok := user.Account.(*MemoryAccount)
	if !ok {
		return newError("user account is not valid")
	}

	iConn := conn
	if statConn, ok := iConn.(*internet.StatCouterConnection); ok {
		iConn = statConn.Connection
	}

	var rawConn syscall.RawConn

	connWriter := &ConnWriter{}
	allowUDP443 := false
	switch account.Flow {
	case XRO + "-udp443", XRD + "-udp443":
		allowUDP443 = true
		account.Flow = account.Flow[:16]
		fallthrough
	case XRO, XRD:
		if destination.Address.Family().IsDomain() && destination.Address.Domain() == muxCoolAddress {
			return newError(account.Flow + " doesn't support Mux").AtWarning()
		}
		if destination.Network == net.Network_UDP {
			if !allowUDP443 && destination.Port == 443 {
				return newError(account.Flow + " stopped UDP/443").AtInfo()
			}
		} else { // enable XTLS only if making TCP request
			if xtlsConn, ok := iConn.(*xtls.Conn); ok {
				xtlsConn.RPRX = true
				xtlsConn.SHOW = trojanXTLSShow
				connWriter.Flow = account.Flow
				if account.Flow == XRD {
					xtlsConn.DirectMode = true
					if sc, ok := xtlsConn.Connection.(syscall.Conn); ok {
						rawConn, _ = sc.SyscallConn()
					}
				}
			} else {
				return newError(`failed to use ` + account.Flow + `, maybe "security" is not "xtls"`).AtWarning()
			}
		}
	case "":
		if _, ok := iConn.(*xtls.Conn); ok {
			panic(`To avoid misunderstanding, you must fill in Trojan "flow" when using XTLS.`)
		}
	default:
		return newError("unsupported flow " + account.Flow).AtWarning()
	}

	sessionPolicy := c.policyManager.ForLevel(user.Level)
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

	postRequest := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		var bodyWriter buf.Writer
		bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		connWriter.Writer = bufferWriter
		connWriter.Target = destination
		connWriter.Account = account

		if destination.Network == net.Network_UDP {
			bodyWriter = &PacketWriter{Writer: connWriter, Target: destination}
		} else {
			bodyWriter = connWriter
		}

		// write some request payload to buffer
		if err = buf.CopyOnceTimeout(link.Reader, bodyWriter, time.Millisecond*100); err != nil && err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
			return newError("failed to write A request payload").Base(err).AtWarning()
		}

		// Flush; bufferWriter.WriteMultiBufer now is bufferWriter.writer.WriteMultiBuffer
		if err = bufferWriter.SetBuffered(false); err != nil {
			return newError("failed to flush payload").Base(err).AtWarning()
		}

		if err = buf.Copy(link.Reader, bodyWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transfer request payload").Base(err).AtInfo()
		}

		return nil
	}

	getResponse := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		var reader buf.Reader
		if network == net.Network_UDP {
			reader = &PacketReader{
				Reader: conn,
			}
		} else {
			reader = buf.NewReader(conn)
		}
		if rawConn != nil {
			return ReadV(reader, link.Writer, timer, iConn.(*xtls.Conn), rawConn)
		}
		return buf.Copy(reader, link.Writer, buf.UpdateActivity(timer))
	}

	var responseDoneAndCloseWriter = task.OnSuccess(getResponse, task.Close(link.Writer))
	if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))

	const defaultFlagValue = "NOT_DEFINED_AT_ALL"

	xtlsShow := platform.NewEnvFlag("v2ray.trojan.xtls.show").GetValue(func() string { return defaultFlagValue })
	if xtlsShow == "true" {
		trojanXTLSShow = true
	}
}
