package stun

import (
	"sync"
	"time"

	"github.com/pion/stun/v3"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type stunTransactionID = [stun.TransactionIDSize]byte

type pendingTransaction struct {
	handler   PendingTransactionHandler
	createdAt time.Time
}

type Processor struct {
	pendingStunRequest map[stunTransactionID]pendingTransaction
	closed             bool
	mux                sync.Mutex
}

func NewProcessor() *Processor {
	return &Processor{
		pendingStunRequest: make(map[stunTransactionID]pendingTransaction),
	}
}

func (p *Processor) HandleStunPacket(b []byte, addr net.Addr) {
	var msg stun.Message
	if err := stun.Decode(b, &msg); err != nil {
		return
	}

	p.mux.Lock()
	pt, ok := p.pendingStunRequest[msg.TransactionID]
	if ok {
		delete(p.pendingStunRequest, msg.TransactionID)
	}
	p.mux.Unlock()

	if ok {
		pt.handler(msg.TransactionID, msg, addr)
	}
}

type PendingTransactionHandler func(transactionID [stun.TransactionIDSize]byte, msg stun.Message, addr net.Addr)

func (p *Processor) AddPendingTransactionListener(transactionID [stun.TransactionIDSize]byte, handler PendingTransactionHandler) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.pendingStunRequest[transactionID] = pendingTransaction{
		handler:   handler,
		createdAt: time.Now(),
	}
}

func (p *Processor) CancelTransaction(transactionID [stun.TransactionIDSize]byte) {
	p.mux.Lock()
	defer p.mux.Unlock()
	delete(p.pendingStunRequest, transactionID)
}

func (p *Processor) ExpiredTransaction(newerThanThisTimeOrExpire time.Time) int {
	p.mux.Lock()
	defer p.mux.Unlock()
	expired := 0
	for id, pt := range p.pendingStunRequest {
		if pt.createdAt.Before(newerThanThisTimeOrExpire) {
			delete(p.pendingStunRequest, id)
			expired++
		}
	}
	return expired
}
