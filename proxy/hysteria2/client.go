package hysteria2

import (
	"context"

	hyProtocol "github.com/v2fly/hysteria/core/v2/international/protocol"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/retry"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/proxy"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	hyTransport "github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

// Client is an inbound handler
type Client struct {
	serverPicker  protocol.ServerPicker
	policyManager policy.Manager
}

// NewClient create a new client.
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
	newError("tunneling request to ", destination, " via ", server.Destination().NetAddr()).WriteToLog(session.ExportIDToError(ctx))

	defer conn.Close()

	iConn := conn
	if statConn, ok := conn.(*internet.StatCouterConnection); ok {
		iConn = statConn.Connection // will not count the UDP traffic.
	}
	hyConn, IsHy2Transport := iConn.(*hyTransport.HyConn)

	if !IsHy2Transport && network == net.Network_UDP {
		// hysteria2 need to use udp extension to proxy UDP.
		return newError(hyTransport.CanNotUseUDPExtension)
	}

	user := server.PickUser()
	userLevel := uint32(0)
	if user != nil {
		userLevel = user.Level
	}
	sessionPolicy := c.policyManager.ForLevel(userLevel)
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

	if packetConn, err := packetaddr.ToPacketAddrConn(link, destination); err == nil {
		postRequest := func() error {
			defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

			var buffer [2048]byte
			n, addr, err := packetConn.ReadFrom(buffer[:])
			if err != nil {
				return newError("failed to read a packet").Base(err)
			}
			dest := net.DestinationFromAddr(addr)

			bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
			connWriter := &ConnWriter{Writer: bufferWriter, Target: dest}
			packetWriter := &PacketWriter{Writer: connWriter, Target: dest, HyConn: hyConn}

			// write some request payload to buffer
			if _, err := packetWriter.WriteTo(buffer[:n], addr); err != nil {
				return newError("failed to write a request payload").Base(err)
			}

			// Flush; bufferWriter.WriteMultiBuffer now is bufferWriter.writer.WriteMultiBuffer
			if err = bufferWriter.SetBuffered(false); err != nil {
				return newError("failed to flush payload").Base(err).AtWarning()
			}

			return udp.CopyPacketConn(packetWriter, packetConn, udp.UpdateActivity(timer))
		}

		getResponse := func() error {
			defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

			packetReader := &PacketReader{Reader: conn, HyConn: hyConn}
			packetConnectionReader := &PacketConnectionReader{reader: packetReader}

			return udp.CopyPacketConn(packetConn, packetConnectionReader, udp.UpdateActivity(timer))
		}

		responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
		if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
			return newError("connection ends").Base(err)
		}

		return nil
	}

	postRequest := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		var bodyWriter buf.Writer
		bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		connWriter := &ConnWriter{Writer: bufferWriter, Target: destination}
		bodyWriter = connWriter

		if network == net.Network_UDP {
			bodyWriter = &PacketWriter{Writer: connWriter, Target: destination, HyConn: hyConn}
		} else {
			// write some request payload to buffer
			err = buf.CopyOnceTimeout(link.Reader, bodyWriter, proxy.FirstPayloadTimeout)
			switch err {
			case buf.ErrNotTimeoutReader, buf.ErrReadTimeout:
				if err := connWriter.WriteTCPHeader(); err != nil {
					return newError("failed to write request header").Base(err).AtWarning()
				}
			case nil:
			default:
				return newError("failed to write a request payload").Base(err).AtWarning()
			}
			// Flush; bufferWriter.WriteMultiBuffer now is bufferWriter.writer.WriteMultiBuffer
			if err = bufferWriter.SetBuffered(false); err != nil {
				return newError("failed to flush payload").Base(err).AtWarning()
			}
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
				Reader: conn, HyConn: hyConn,
			}
		} else {
			ok, msg, err := hyProtocol.ReadTCPResponse(conn)
			if err != nil {
				return err
			}
			if !ok {
				return newError(msg)
			}
			reader = buf.NewReader(conn)
		}
		return buf.Copy(reader, link.Writer, buf.UpdateActivity(timer))
	}

	responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
	if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
