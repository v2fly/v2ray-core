package rrpitBidirectionalSession

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitMaterializedTransferChannel"
)

type BidirectionalSession struct {
	mu   sync.Mutex
	cond *sync.Cond

	rx *rriptMonoDirectionSession.SessionRx
	tx *rriptMonoDirectionSession.SessionTx

	TimestampInterval time.Duration

	nextTimestamp uint64
	autoTickStop  chan struct{}
	autoTickDone  chan struct{}
	closeOnce     sync.Once
	closed        bool
}

type Config struct {
	Rx rriptMonoDirectionSession.SessionRxConfig
	Tx rriptMonoDirectionSession.SessionTxConfig

	TimestampInterval time.Duration
}

func New(config Config) (*BidirectionalSession, error) {
	session := &BidirectionalSession{
		TimestampInterval: config.TimestampInterval,
	}
	session.cond = sync.NewCond(&session.mu)

	tx, err := rriptMonoDirectionSession.NewSessionTx(config.Tx)
	if err != nil {
		return nil, err
	}
	session.tx = tx

	rxConfig := config.Rx
	userRemoteControlHandler := rxConfig.OnRemoteControlMsg
	rxConfig.OnRemoteControlMsg = func(ctrl rriptMonoDirectionSession.ControlMessage) error {
		session.mu.Lock()
		if session.closed {
			session.mu.Unlock()
			return io.ErrClosedPipe
		}
		if err := tx.AcceptRemoteControlMessage(ctrl); err != nil {
			session.mu.Unlock()
			return err
		}
		if session.cond != nil {
			session.cond.Broadcast()
		}
		session.mu.Unlock()
		if userRemoteControlHandler != nil {
			return userRemoteControlHandler(ctrl)
		}
		return nil
	}

	rx, err := rriptMonoDirectionSession.NewSessionRx(rxConfig)
	if err != nil {
		return nil, err
	}
	session.rx = rx
	session.startAutoTick()
	return session, nil
}

func (s *BidirectionalSession) SendMessage(data []byte) error {
	if s == nil || s.tx == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		if s.closed {
			return io.ErrClosedPipe
		}
		err := s.tx.SendMessage(data)
		if err == nil {
			return nil
		}
		if !errors.Is(err, rriptMonoDirectionSession.ErrTxLaneBufferFull) {
			return err
		}
		if s.cond == nil {
			return err
		}
		s.cond.Wait()
	}
}

func (s *BidirectionalSession) OnNewTimestamp(timestamp uint64) error {
	if s == nil || s.tx == nil || s.rx == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if timestamp > s.nextTimestamp {
		s.nextTimestamp = timestamp
	}
	return s.onNewTimestampLocked(timestamp)
}

func (s *BidirectionalSession) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		if s.cond != nil {
			s.cond.Broadcast()
		}
		s.mu.Unlock()
		if s.autoTickStop != nil {
			close(s.autoTickStop)
			<-s.autoTickDone
		}
	})
	return nil
}

func (s *BidirectionalSession) onNewTimestampLocked(timestamp uint64) error {
	if err := s.tx.OnNewTimestamp(timestamp); err != nil {
		return err
	}
	return s.tx.FloodControlMessages(s.rx.GenerateControlMessage)
}

func (s *BidirectionalSession) AttachTxChannel(closer io.WriteCloser) (channelID uint64, err error) {
	return s.AttachTxChannelWithConfig(closer, rriptMonoDirectionSession.ChannelConfig{Weight: 1})
}

func (s *BidirectionalSession) AttachTxChannelWithConfig(
	closer io.WriteCloser,
	config rriptMonoDirectionSession.ChannelConfig,
) (channelID uint64, err error) {
	if s == nil || s.tx == nil || s.rx == nil {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	channelID, err = s.tx.AttachTxChannelWithConfig(closer, config)
	if err != nil {
		return 0, err
	}
	return channelID, nil
}

func (s *BidirectionalSession) AttachRxChannel() (rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	if s == nil || s.rx == nil {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	rxChannel, err = s.rx.AttachRxChannel(0)
	if err != nil {
		return nil, err
	}
	return rxChannel, nil
}

func (s *BidirectionalSession) AttachChannel(closer io.WriteCloser) (channelID uint64, rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	return s.AttachChannelWithConfig(closer, rriptMonoDirectionSession.ChannelConfig{Weight: 1})
}

func (s *BidirectionalSession) AttachChannelWithConfig(
	closer io.WriteCloser,
	config rriptMonoDirectionSession.ChannelConfig,
) (channelID uint64, rxChannel *rrpitMaterializedTransferChannel.ChannelRx, err error) {
	if s == nil {
		return 0, nil, nil
	}
	channelID, err = s.AttachTxChannelWithConfig(closer, config)
	if err != nil {
		return 0, nil, err
	}
	rxChannel, err = s.AttachRxChannel()
	if err != nil {
		return 0, nil, err
	}
	return channelID, rxChannel, nil
}

func (s *BidirectionalSession) Rx() *rriptMonoDirectionSession.SessionRx {
	if s == nil {
		return nil
	}
	return s.rx
}

func (s *BidirectionalSession) Tx() *rriptMonoDirectionSession.SessionTx {
	if s == nil {
		return nil
	}
	return s.tx
}

func (s *BidirectionalSession) MaxMessageSize() (int, error) {
	if s == nil || s.tx == nil {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tx.MaxMessageSize()
}

func (s *BidirectionalSession) startAutoTick() {
	if s == nil || s.TimestampInterval <= 0 {
		return
	}

	s.autoTickStop = make(chan struct{})
	s.autoTickDone = make(chan struct{})
	go func() {
		ticker := time.NewTicker(s.TimestampInterval)
		defer func() {
			ticker.Stop()
			close(s.autoTickDone)
		}()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				s.nextTimestamp += 1
				err := s.onNewTimestampLocked(s.nextTimestamp)
				s.mu.Unlock()
				if err != nil {
					return
				}
			case <-s.autoTickStop:
				return
			}
		}
	}()
}
