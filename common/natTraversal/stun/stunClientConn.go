package stun

import (
	"errors"
	"time"

	"github.com/pion/stun/v3"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

var ErrTimeout = errors.New("STUN transaction timed out")

func NewStunClientConn(conn net.PacketConn) (*StunClientConn, error) {
	processor := NewProcessor()
	filtered, err := NewFilteredConnection(conn, processor.HandleStunPacket)
	if err != nil {
		return nil, err
	}
	return &StunClientConn{
		PacketConn: filtered,
		processor:  processor,
	}, nil
}

type StunClientConn struct {
	net.PacketConn
	processor *Processor
}

func (conn *StunClientConn) ExecuteSTUNMessage(msg stun.Message, dest net.Addr, timeout time.Duration) (resp stun.Message, addr net.Addr, err error) {
	type result struct {
		msg  stun.Message
		addr net.Addr
	}
	ch := make(chan result, 1)

	_, _, err = conn.ExecuteSTUNMessageAsync(msg, dest, func(_ [stun.TransactionIDSize]byte, respMsg stun.Message, respAddr net.Addr) {
		ch <- result{msg: respMsg, addr: respAddr}
	})
	if err != nil {
		return resp, nil, err
	}

	select {
	case r := <-ch:
		return r.msg, r.addr, nil
	case <-time.After(timeout):
		conn.processor.CancelTransaction(msg.TransactionID)
		return resp, nil, ErrTimeout
	}
}

func (conn *StunClientConn) ExecuteSTUNMessageAsync(msg stun.Message, dest net.Addr, callback PendingTransactionHandler) (resp stun.Message, addr net.Addr, err error) {
	msg.Encode()
	conn.processor.AddPendingTransactionListener(msg.TransactionID, callback)

	if _, err = conn.WriteTo(msg.Raw, dest); err != nil {
		conn.processor.CancelTransaction(msg.TransactionID)
		return resp, nil, err
	}

	return resp, nil, nil
}
