package hysteria2

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/apernet/quic-go"
	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}

// Server is an inbound connection handler that handles messages in protocol.
type Server struct {
	policyManager  policy.Manager
	packetEncoding packetaddr.PacketAddrType
}

// NewServer creates a new inbound handler.
func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {
	v := core.MustFromContext(ctx)
	server := &Server{
		policyManager:  v.GetFeature(policy.ManagerType()).(policy.Manager),
		packetEncoding: config.PacketEncoding,
	}
	return server, nil
}

// Network implements proxy.Inbound.Network().
func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

// Process implements proxy.Inbound.Process().
func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	sid := session.ExportIDToError(ctx)
	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		panic("no inbound metadata")
	}

	iConn := conn
	if statConn, ok := conn.(*internet.StatCouterConnection); ok {
		iConn = statConn.Connection
	}
	if _, ok := iConn.(*hysteria2.InterConn); ok {
		network = net.Network_UDP
	}

	if network == net.Network_UDP {
		udpDispatcherConstructor := udp.NewSplitDispatcher
		switch s.packetEncoding {
		case packetaddr.PacketAddrType_None:
		case packetaddr.PacketAddrType_Packet:
			packetAddrDispatcherFactory := udp.NewPacketAddrDispatcherCreator(ctx)
			udpDispatcherConstructor = packetAddrDispatcherFactory.NewPacketAddrDispatcher
		}

		var writeBuf [buf.Size]byte
		udpServer := udpDispatcherConstructor(dispatcher, func(ctx context.Context, packet *udp_proto.Packet) {
			msg := &UDPMessage{
				SessionID: 0,
				PacketID:  0,
				FragID:    0,
				FragCount: 1,
				Addr:      packet.Source.NetAddr(),
				Data:      packet.Payload.Bytes(),
			}
			msgN := msg.Serialize(writeBuf[:])
			_, err := conn.Write(writeBuf[:msgN])
			var errTooLarge *quic.DatagramTooLargeError
			if errors.As(err, &errTooLarge) {
				msg.PacketID = uint16(rand.Intn(0xFFFF)) + 1
				fMsgs := FragUDPMessage(msg, int(errTooLarge.MaxDatagramPayloadSize))
				for _, fMsg := range fMsgs {
					msgN = fMsg.Serialize(writeBuf[:])
					_, err = conn.Write(writeBuf[:msgN])
					if err != nil {
						break
					}
				}
			}
			packet.Payload.Release()
		})

		var df = &Defragger{}
		for {
			var readBuf [hysteria2.MaxDatagramFrameSize]byte

			n, err := conn.Read(readBuf[:])
			if err != nil {
				return err
			}

			msg, err := ParseUDPMessage(readBuf[:n])
			if err != nil {
				continue
			}

			dfMsg := df.Feed(msg)
			if dfMsg == nil {
				continue
			}

			destination, err := net.ParseDestination("udp:" + dfMsg.Addr)
			if err != nil {
				continue
			}

			payload := buf.New()
			if _, err := payload.Write(dfMsg.Data); err != nil {
				payload.Release()
				continue
			}

			currentPacketCtx := ctx
			currentPacketCtx = log.ContextWithAccessMessage(currentPacketCtx, &log.AccessMessage{
				From:   inbound.Source,
				To:     destination,
				Status: log.AccessAccepted,
				Reason: "",
			})
			newError("tunnelling request to ", destination).WriteToLog(sid)

			udpServer.Dispatch(currentPacketCtx, destination, payload)
		}
	}

	sessionPolicy := s.policyManager.ForLevel(0)
	if err := conn.SetReadDeadline(time.Now().Add(sessionPolicy.Timeouts.Handshake)); err != nil {
		return newError("unable to set read deadline").Base(err).AtWarning()
	}

	addr, err := ReadTCPRequest(conn)
	if err != nil {
		log.Record(&log.AccessMessage{
			From:   conn.RemoteAddr(),
			To:     "",
			Status: log.AccessRejected,
			Reason: err,
		})
		return newError("failed to create request from: ", conn.RemoteAddr()).Base(err)
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return newError("unable to set read deadline").Base(err).AtWarning()
	}

	destination, err := net.ParseDestination("tcp:" + addr)
	if err != nil {
		return err
	}

	ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   conn.RemoteAddr(),
		To:     destination,
		Status: log.AccessAccepted,
		Reason: "",
	})

	newError("received request for ", destination).WriteToLog(sid)

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)

	link, err := dispatcher.Dispatch(ctx, destination)
	if err != nil {
		return newError("failed to dispatch request to ", destination).Base(err)
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transfer request").Base(err)
		}

		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		err := WriteTCPResponse(bufferedWriter, true, "")
		if err != nil {
			return newError("failed to write response").Base(err).AtWarning()
		}
		if err = bufferedWriter.SetBuffered(false); err != nil {
			return newError("failed to flush payload").Base(err).AtWarning()
		}
		if err := buf.Copy(link.Reader, bufferedWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport response").Base(err)
		}

		return nil
	}

	requestDonePost := task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDonePost, responseDone); err != nil {
		common.Must(common.Interrupt(link.Reader))
		common.Must(common.Interrupt(link.Writer))
		return newError("connection ends").Base(err)
	}

	return nil
}
