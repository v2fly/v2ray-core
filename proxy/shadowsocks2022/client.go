package shadowsocks2022

import (
	"context"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/retry"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"time"
)

type Client struct {
	config *ClientConfig
	ctx    context.Context
}

func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified")
	}
	destination := outbound.Target
	network := destination.Network

	var conn internet.Connection

	err := retry.ExponentialBackoff(5, 100).On(func() error {
		dest := net.TCPDestination(c.config.Address.AsAddress(), net.Port(c.config.Port))
		dest.Network = network
		rawConn, err := dialer.Dial(ctx, dest)
		if err != nil {
			return err
		}
		conn = rawConn

		return nil
	})
	if err != nil {
		return newError("failed to find an available destination").AtWarning().Base(err)
	}
	newError("tunneling request to ", destination, " via ", network, ":", c.config.Address).WriteToLog(session.ExportIDToError(ctx))
	defer conn.Close()

	var keyDerivation = newBLAKE3KeyDerivation()
	var method Method
	switch c.config.Method {
	case "2022-blake3-aes-128-gcm":
		method = newAES128GCMMethod()
	case "2022-blake3-aes-256-gcm":
		method = newAES256GCMMethod()
	default:
		return newError("unknown method: ", c.config.Method)
	}

	effectivePsk := c.config.Psk

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, time.Minute)

	if network == net.Network_TCP {
		request := &TCPRequest{
			keyDerivation: keyDerivation,
			method:        method,
		}
		TCPRequestBuffer := buf.New()
		defer TCPRequestBuffer.Release()
		err = request.EncodeTCPRequestHeader(effectivePsk, c.config.Ipsk, destination.Address,
			int(destination.Port), nil, TCPRequestBuffer)
		if err != nil {
			return newError("failed to encode TCP request header").Base(err)
		}
		_, err = conn.Write(TCPRequestBuffer.Bytes())
		if err != nil {
			return newError("failed to write TCP request header").Base(err)
		}
		requestDone := func() error {
			encodedWriter := request.CreateClientC2SWriter(conn)
			return buf.Copy(link.Reader, encodedWriter, buf.UpdateActivity(timer))
		}
		responseDone := func() error {
			err = request.DecodeTCPResponseHeader(effectivePsk, conn)
			if err != nil {
				return newError("failed to decode TCP response header").Base(err)
			}
			if err = request.CheckC2SConnectionConstraint(); err != nil {
				return newError("C2S connection constraint violation").Base(err)
			}
			initialPayload := buf.NewWithSize(65535)
			encodedReader, err := request.CreateClientS2CReader(conn, initialPayload)
			if err != nil {
				return newError("failed to create client S2C reader").Base(err)
			}
			err = link.Writer.WriteMultiBuffer(buf.MultiBuffer{initialPayload})
			if err != nil {
				return newError("failed to write initial payload").Base(err)
			}
			return buf.Copy(encodedReader, link.Writer, buf.UpdateActivity(timer))
		}
		responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
		if err := task.Run(ctx, requestDone, responseDoneAndCloseWriter); err != nil {
			return newError("connection ends").Base(err)
		}
		return nil
	} else {
		return newError("not implemented")
	}
}

func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	return &Client{
		config: config,
		ctx:    ctx,
	}, nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		clientConfig, ok := config.(*ClientConfig)
		if !ok {
			return nil, newError("not a ClientConfig")
		}
		return NewClient(ctx, clientConfig)
	}))
}
