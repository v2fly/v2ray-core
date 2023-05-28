package tun

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/log"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const (
	rcvWnd      = 0 // default settings
	maxInFlight = 2 << 10
)

type tcpConn struct {
	*gonet.TCPConn
	id stack.TransportEndpointID
}

func (c *tcpConn) ID() *stack.TransportEndpointID {
	return &c.id
}

type TCPHandler struct {
	ctx           context.Context
	dispatcher    routing.Dispatcher
	policyManager policy.Manager
	config        *Config

	stack *stack.Stack
}

func HandleTCP(handle func(TCPConn)) StackOption {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, rcvWnd, maxInFlight, func(r *tcp.ForwarderRequest) {
			wg := new(waiter.Queue)
			linkedEndpoint, err := r.CreateEndpoint(wg)
			if err != nil {
				r.Complete(true)
				return
			}
			defer r.Complete(false)

			// TODO: set sockopt

			// tcpHandler := &TCPHandler{
			// 	ctx:           ctx,
			// 	dispatcher:    dispatcher,
			// 	policyManager: policyManager,
			// 	config:        config,
			// 	stack:         s,
			// }

			// if err := tcpHandler.Handle(gonet.NewTCPConn(wg, linkedEndpoint)); err != nil {
			// 	// TODO: log
			// 	// return newError("failed to handle tcp connection").Base(err)
			// }

			tcpConn := &tcpConn{
				TCPConn: gonet.NewTCPConn(wg, linkedEndpoint),
				id:      r.ID(),
			}

			tcpQueue <- tcpConn
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

		return nil
	}
}

func (h *TCPHandler) HandleQueue(ch chan TCPConn) {
	for {
		select {
		case conn := <-ch:
			if err := h.Handle(conn); err != nil {
				newError(err).AtError().WriteToLog(session.ExportIDToError(h.ctx))
			}
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *TCPHandler) Handle(conn TCPConn) error {
	ctx := session.ContextWithInbound(h.ctx, &session.Inbound{Tag: h.config.Tag})
	sessionPolicy := h.policyManager.ForLevel(h.config.UserLevel)

	addr := conn.RemoteAddr()

	dest := net.DestinationFromAddr(addr)
	ctx = log.ContextWithAccessMessage(h.ctx, &log.AccessMessage{
		From:   addr,
		To:     dest,
		Status: log.AccessAccepted,
		Reason: "",
	})
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	link, err := h.dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return newError("failed to dispatch").Base(err)
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		if err := buf.Copy(link.Reader, buf.NewWriter(conn), buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP response").Base(err)
		}

		return nil
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transport all TCP request").Base(err)
		}

		return nil
	}

	requestDoneAndCloseWriter := task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(h.ctx, requestDoneAndCloseWriter, responseDone); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}

	return nil
}