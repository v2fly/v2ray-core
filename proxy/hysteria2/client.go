package hysteria2

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"strconv"

	"github.com/apernet/quic-go"
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
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"
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
		rawConn, err := dialer.Dial(hysteria2.ContextWithDatagram(ctx, network == net.Network_UDP), server.Destination())
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
		iConn = statConn.Connection
	}
	if _, ok := iConn.(*hysteria2.InterConn); !ok && network == net.Network_UDP {
		return newError("udp require hysteria2 transport")
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

			return udp.CopyPacketConn(&UDPWriter{writer: conn}, packetConn, udp.UpdateActivity(timer))
		}

		getResponse := func() error {
			defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

			return udp.CopyPacketConn(packetConn, &UDPReader{reader: conn, df: &Defragger{}}, udp.UpdateActivity(timer))
		}

		responseDoneAndCloseWriter := task.OnSuccess(getResponse, task.Close(link.Writer))
		if err := task.Run(ctx, postRequest, responseDoneAndCloseWriter); err != nil {
			return newError("connection ends").Base(err)
		}

		return nil
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		var bodyWriter buf.Writer

		if network == net.Network_UDP {
			bodyWriter = &UDPWriter{writer: conn, addr: destination.NetAddr()}
		} else {
			bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
			bodyWriter = bufferedWriter
			err := WriteTCPRequest(bufferedWriter, destination.NetAddr())
			if err != nil {
				return newError("failed to write request").Base(err).AtWarning()
			}
			if err = buf.CopyOnceTimeout(link.Reader, bufferedWriter, proxy.FirstPayloadTimeout); err != nil && err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
				return newError("failed to write request payload").Base(err).AtWarning()
			}
			if err = bufferedWriter.SetBuffered(false); err != nil {
				return newError("failed to flush payload").Base(err).AtWarning()
			}
		}

		return buf.Copy(link.Reader, bodyWriter, buf.UpdateActivity(timer))
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		var reader buf.Reader

		if network == net.Network_UDP {
			reader = &UDPReader{reader: conn, df: &Defragger{}}
		} else {
			ok, msg, err := ReadTCPResponse(conn)
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

	responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDone, responseDoneAndCloseWriter); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}

type UDPWriter struct {
	writer io.Writer
	addr   string
	buf    [buf.Size]byte
}

func (w *UDPWriter) Network() string {
	return "udp"
}

func (w *UDPWriter) String() string {
	return w.addr
}

func (w *UDPWriter) SendMessage(msg *UDPMessage) error {
	msgN := msg.Serialize(w.buf[:])
	if msgN < 0 {
		return nil
	}
	_, err := w.writer.Write(w.buf[:msgN])
	return err
}

func (w *UDPWriter) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	msg := &UDPMessage{
		SessionID: 0,
		PacketID:  0,
		FragID:    0,
		FragCount: 1,
		Addr:      addr.String(),
		Data:      p,
	}
	err = w.SendMessage(msg)
	var errTooLarge *quic.DatagramTooLargeError
	if errors.As(err, &errTooLarge) {
		msg.PacketID = uint16(rand.Intn(0xFFFF)) + 1
		fMsgs := FragUDPMessage(msg, int(errTooLarge.MaxDatagramPayloadSize))
		for _, fMsg := range fMsgs {
			err = w.SendMessage(&fMsg)
			if err != nil {
				return 0, err
			}
		}
	} else if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *UDPWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for i, b := range mb {
		_, err := w.WriteTo(b.Bytes(), w)
		if err != nil {
			buf.ReleaseMulti(mb[i:])
			return err
		}
		b.Release()
	}
	return nil
}

type UDPReader struct {
	reader io.Reader
	df     *Defragger
}

func (r *UDPReader) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	for {
		var buf [hysteria2.MaxDatagramFrameSize]byte

		n, err := r.reader.Read(buf[:])
		if err != nil {
			return 0, nil, err
		}

		msg, err := ParseUDPMessage(buf[:n])
		if err != nil {
			continue
		}

		dfMsg := r.df.Feed(msg)
		if dfMsg == nil {
			continue
		}

		if len(p) < len(dfMsg.Data) {
			continue
		}

		host, port, _ := net.SplitHostPort(dfMsg.Addr)
		ip := net.ParseIP(host)
		if ip == nil {
			continue
		}
		portint, _ := strconv.Atoi(port)

		return copy(p, dfMsg.Data), &net.UDPAddr{IP: ip, Port: portint}, nil
	}
}

func (r *UDPReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	b := buf.New()
	b.Resize(0, buf.Size)
	n, _, err := r.ReadFrom(b.Bytes())
	if err != nil {
		b.Release()
		return nil, err
	}
	b.Resize(0, int32(n))
	return buf.MultiBuffer{b}, nil
}
