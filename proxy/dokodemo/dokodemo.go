package dokodemo

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"sync/atomic"
	"time"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/log"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/common/signal"
	"github.com/ghxhy/v2ray-core/v5/common/task"
	"github.com/ghxhy/v2ray-core/v5/features/policy"
	"github.com/ghxhy/v2ray-core/v5/features/routing"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		d := new(Door)
		err := core.RequireFeatures(ctx, func(pm policy.Manager) error {
			return d.Init(config.(*Config), pm, session.SockoptFromContext(ctx))
		})
		return d, err
	}))

	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedServer := config.(*SimplifiedConfig)
		fullConfig := &Config{
			Address:        simplifiedServer.Address,
			Port:           simplifiedServer.Port,
			Networks:       simplifiedServer.Networks.Network,
			FollowRedirect: simplifiedServer.FollowRedirect,
		}

		return common.CreateObject(ctx, fullConfig)
	}))
}

type Door struct {
	policyManager policy.Manager
	config        *Config
	address       net.Address
	port          net.Port
	sockopt       *session.Sockopt
}

// Init initializes the Door instance with necessary parameters.
func (d *Door) Init(config *Config, pm policy.Manager, sockopt *session.Sockopt) error {
	if (config.NetworkList == nil || len(config.NetworkList.Network) == 0) && len(config.Networks) == 0 {
		return newError("no network specified")
	}
	d.config = config
	d.address = config.GetPredefinedAddress()
	d.port = net.Port(config.Port)
	d.policyManager = pm
	d.sockopt = sockopt

	return nil
}

// Network implements proxy.Inbound.
func (d *Door) Network() []net.Network {
	if len(d.config.Networks) > 0 {
		return d.config.Networks
	}

	return d.config.NetworkList.GetNetwork()
}

func (d *Door) policy() policy.Session {
	config := d.config
	p := d.policyManager.ForLevel(config.UserLevel)
	if config.Timeout > 0 && config.UserLevel == 0 {
		p.Timeouts.ConnectionIdle = time.Duration(config.Timeout) * time.Second
	}
	return p
}

type hasHandshakeAddress interface {
	HandshakeAddress() net.Address
}

// Process implements proxy.Inbound.
func (d *Door) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	newError("processing connection from: ", conn.RemoteAddr()).AtDebug().WriteToLog(session.ExportIDToError(ctx))
	dest := net.Destination{
		Network: network,
		Address: d.address,
		Port:    d.port,
	}

	destinationOverridden := false
	if d.config.FollowRedirect {
		if outbound := session.OutboundFromContext(ctx); outbound != nil && outbound.Target.IsValid() {
			dest = outbound.Target
			destinationOverridden = true
		} else if handshake, ok := conn.(hasHandshakeAddress); ok {
			addr := handshake.HandshakeAddress()
			if addr != nil {
				dest.Address = addr
				destinationOverridden = true
			}
		}
	}
	if !dest.IsValid() || dest.Address == nil {
		return newError("unable to get destination")
	}

	if inbound := session.InboundFromContext(ctx); inbound != nil {
		inbound.User = &protocol.MemoryUser{
			Level: d.config.UserLevel,
		}
	}

	ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   conn.RemoteAddr(),
		To:     dest,
		Status: log.AccessAccepted,
		Reason: "",
	})
	newError("received request for ", conn.RemoteAddr()).WriteToLog(session.ExportIDToError(ctx))

	plcy := d.policy()
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, plcy.Timeouts.ConnectionIdle)

	ctx = policy.ContextWithBufferPolicy(ctx, plcy.Buffer)
	link, err := dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return newError("failed to dispatch request").Base(err)
	}

	requestCount := int32(1)
	requestDone := func() error {
		defer func() {
			if atomic.AddInt32(&requestCount, -1) == 0 {
				timer.SetTimeout(plcy.Timeouts.DownlinkOnly)
			}
		}()

		var reader buf.Reader
		if dest.Network == net.Network_UDP {
			reader = buf.NewPacketReader(conn)
		} else {
			reader = buf.NewReader(conn)
		}
		if err := buf.Copy(reader, link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport request").Base(err)
		}

		return nil
	}

	tproxyRequest := func() error {
		return nil
	}

	var writer buf.Writer
	if network == net.Network_TCP {
		writer = buf.NewWriter(conn)
	} else {
		// if we are in TPROXY mode, use linux's udp forging functionality
		if !destinationOverridden {
			writer = &buf.SequentialWriter{Writer: conn}
		} else {
			sockopt := &internet.SocketConfig{
				Tproxy: internet.SocketConfig_TProxy,
			}
			if dest.Address.Family().IsIP() {
				sockopt.BindAddress = dest.Address.IP()
				sockopt.BindPort = uint32(dest.Port)
			}
			if d.sockopt != nil {
				sockopt.Mark = d.sockopt.Mark
			}
			tConn, err := internet.DialSystem(ctx, net.DestinationFromAddr(conn.RemoteAddr()), sockopt)
			if err != nil {
				return err
			}
			defer tConn.Close()

			writer = &buf.SequentialWriter{Writer: tConn}
			tReader := buf.NewPacketReader(tConn)
			requestCount++
			tproxyRequest = func() error {
				defer func() {
					if atomic.AddInt32(&requestCount, -1) == 0 {
						timer.SetTimeout(plcy.Timeouts.DownlinkOnly)
					}
				}()
				if err := buf.Copy(tReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
					return newError("failed to transport request (TPROXY conn)").Base(err)
				}
				return nil
			}
		}
	}

	responseDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.UplinkOnly)

		if err := buf.Copy(link.Reader, writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport response").Base(err)
		}
		return nil
	}

	if err := task.Run(ctx, task.OnSuccess(func() error {
		return task.Run(ctx, requestDone, tproxyRequest)
	}, task.Close(link.Writer)), responseDone); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}

	return nil
}
