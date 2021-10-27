package wireguard

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/common/signal"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/common/task"
	"github.com/v2fly/v2ray-core/v4/features/dns"
	"github.com/v2fly/v2ray-core/v4/features/policy"
	"github.com/v2fly/v2ray-core/v4/features/routing"
	"github.com/v2fly/v2ray-core/v4/proxy"
	"github.com/v2fly/v2ray-core/v4/transport"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"strings"
	"sync"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		o := new(Outbound)
		err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher, policyManager policy.Manager, dnsClient dns.Client) error {
			o.ctx = ctx
			o.dispatcher = dispatcher
			o.dnsClient = dnsClient
			o.init = done.New()
			return o.Init(config.(*Config), policyManager)
		})
		return o, err
	}))
	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		o := new(Outbound)
		err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher, policyManager policy.Manager, dnsClient dns.Client) error {
			sf := config.(*SimplifiedConfig)
			cf := &Config{
				Server: &protocol.ServerEndpoint{
					Address: sf.Address,
					Port:    sf.Port,
				},
				Network:       sf.Network,
				PrivateKey:    sf.PrivateKey,
				PeerPublicKey: sf.PeerPublicKey,
				PreSharedKey:  sf.PreSharedKey,
				Mtu:           sf.Mtu,
				UserLevel:     sf.UserLevel,
			}
			o.ctx = ctx
			o.dispatcher = dispatcher
			o.dnsClient = dnsClient
			o.init = done.New()
			return o.Init(cf, policyManager)
		})
		return o, err
	}))
}

var _ proxy.Outbound = (*Outbound)(nil)
var _ conn.Bind = (*Outbound)(nil)

type Outbound struct {
	sync.Mutex

	ctx           context.Context
	dispatcher    routing.Dispatcher
	sessionPolicy policy.Session
	dnsClient     dns.Client

	tun    tun.Device
	dev    *device.Device
	wire   *Net
	dialer internet.Dialer

	init        *done.Instance
	destination net.Destination
	endpoint    *conn.StdNetEndpoint
	connection  *remoteConnection
}

func (o *Outbound) Init(config *Config, policyManager policy.Manager) error {
	o.sessionPolicy = policyManager.ForLevel(config.UserLevel)
	spec, err := protocol.NewServerSpecFromPB(config.Server)
	if err != nil {
		return err
	}

	o.destination = spec.Destination()
	o.endpoint = &conn.StdNetEndpoint{
		Port: int(o.destination.Port),
	}

	if o.destination.Address.Family().IsDomain() {
		o.endpoint.IP = []byte{172, 19, 0, 3}
	} else {
		o.endpoint.IP = o.destination.Address.IP()
	}

	localAddress := make([]net.IP, len(config.LocalAddress))
	if len(localAddress) == 0 {
		return newError("empty local address")
	}
	for index, address := range config.LocalAddress {
		localAddress[index] = net.ParseIP(address)
	}

	var privateKey, peerPublicKey, preSharedKey string
	{
		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(config.PrivateKey))
		bytes, err := buf.ReadAllToBytes(decoder)
		if err != nil {
			return newError("failed to decode private key from base64: ", config.PrivateKey).Base(err)
		}
		privateKey = hex.EncodeToString(bytes)
	}
	{
		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(config.PeerPublicKey))
		bytes, err := buf.ReadAllToBytes(decoder)
		if err != nil {
			return newError("failed to decode peer public key from base64: ", config.PeerPublicKey).Base(err)
		}
		peerPublicKey = hex.EncodeToString(bytes)
	}
	if config.PreSharedKey != "" {
		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(config.PreSharedKey))
		bytes, err := buf.ReadAllToBytes(decoder)
		if err != nil {
			return newError("failed to decode pre share key from base64: ", config.PreSharedKey).Base(err)
		}
		preSharedKey = hex.EncodeToString(bytes)
	}
	ipcConf := "private_key=" + privateKey
	ipcConf += "\npublic_key=" + peerPublicKey
	ipcConf += "\nendpoint=" + o.endpoint.DstToString()

	if preSharedKey != "" {
		ipcConf += "\npreshared_key=" + preSharedKey
	}

	var has4, has6 bool

	for _, address := range localAddress {
		if address.To4() != nil {
			has4 = true
		} else {
			has6 = true
		}
	}

	if has4 {
		ipcConf += "\nallowed_ip=0.0.0.0/0"
	}

	if has6 {
		ipcConf += "\nallowed_ip=::/0"
	}

	mtu := int(config.Mtu)
	if mtu == 0 {
		mtu = 1450
	}
	tun, wire, err := CreateNetTUN(localAddress, mtu)

	if err != nil {
		return newError("failed to create wireguard device").Base(err)
	}

	dev := device.NewDevice(tun, o, device.NewLogger(device.LogLevelVerbose, ""))
	err = dev.IpcSet(ipcConf)
	if err != nil {
		return newError("failed to set wireguard ipc conf").Base(err)
	}

	o.tun = tun
	o.dev = dev
	o.wire = wire

	return nil
}

func (o *Outbound) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	if o.dialer == nil {
		o.dialer = dialer
	}
	o.init.Close()

	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified")
	}
	destination := outbound.Target

	if destination.Address.Family().IsDomain() {
		if c, ok := o.dnsClient.(dns.ClientWithIPOption); ok {
			c.SetFakeDNSOption(false)
		}
		ips, err := o.dnsClient.LookupIP(destination.Address.Domain())
		if err != nil {
			return newError("failed to lookup ip addresses for domain ", destination.Address.Domain()).Base(err)
		}
		destination.Address = net.IPAddress(ips[0])
	}

	var conn internet.Connection
	{
		var err error

		switch destination.Network {
		case net.Network_TCP:
			conn, err = o.wire.DialContextTCP(ctx, &net.TCPAddr{
				IP:   destination.Address.IP(),
				Port: int(destination.Port),
			})
		case net.Network_UDP:
			conn, err = o.wire.DialUDP(nil, &net.UDPAddr{
				IP:   destination.Address.IP(),
				Port: int(destination.Port),
			})
		}

		if err != nil {
			return err
		}
	}

	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, o.sessionPolicy.Timeouts.ConnectionIdle)
	ctx = policy.ContextWithBufferPolicy(ctx, o.sessionPolicy.Buffer)

	uplink := func() error {
		defer timer.SetTimeout(o.sessionPolicy.Timeouts.UplinkOnly)

		if err := buf.Copy(link.Reader, buf.NewWriter(conn), buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP response").Base(err)
		}

		return nil
	}

	downlink := func() error {
		defer timer.SetTimeout(o.sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP request").Base(err)
		}

		return nil
	}

	if err := task.Run(ctx, uplink, downlink); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}

	return nil

}

type remoteConnection struct {
	internet.Connection
	done *done.Instance
}

func (r remoteConnection) Close() error {
	if !r.done.Done() {
		r.done.Close()
	}
	return r.Connection.Close()
}

func (o *Outbound) connect() (*remoteConnection, error) {
	if o.dialer == nil {
		<-o.init.Wait()
	}

	if c := o.connection; c != nil && !c.done.Done() {
		return c, nil
	}

	o.Lock()
	defer o.Unlock()

	if c := o.connection; c != nil && !c.done.Done() {
		return c, nil
	}

	conn, err := o.dialer.Dial(context.Background(), o.destination)
	if err == nil {
		o.connection = &remoteConnection{
			conn,
			done.New(),
		}
	}

	return o.connection, err
}

func (o *Outbound) Open(uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	return []conn.ReceiveFunc{o.Receive}, 0, nil
}

func (o *Outbound) Receive(b []byte) (n int, ep conn.Endpoint, err error) {
	var c *remoteConnection
	c, err = o.connect()
	if err != nil {
		return
	}
	n, err = c.Read(b)
	if err != nil {
		common.Close(c)
	} else {
		ep = o.endpoint
	}
	return
}

func (o *Outbound) Close() error {
	o.Lock()
	defer o.Unlock()

	c := o.connection
	if c != nil {
		common.Close(c)
	}

	return nil
}

func (o *Outbound) SetMark(uint32) error {
	return nil
}

func (o *Outbound) Send(b []byte, _ conn.Endpoint) (err error) {
	var c *remoteConnection
	c, err = o.connect()
	if err != nil {
		return
	}
	_, err = c.Write(b)
	if err != nil {
		common.Close(c)
	}
	return err
}

func (o *Outbound) ParseEndpoint(string) (conn.Endpoint, error) {
	return o.endpoint, nil
}
