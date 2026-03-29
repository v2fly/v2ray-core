package rrpitMaterializedTransferChannel

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
)

const channelSequenceFieldLength = 8

type ChannelRx struct {
	rrpitTransferChannel.ChannelRx
	OnNewDataMessage func(data []byte) error

	closed bool
}

func NewChannelRx(channelID uint64, onNewDataMessage func(data []byte) error) (*ChannelRx, error) {
	if onNewDataMessage == nil {
		return nil, newError("nil OnNewDataMessage callback")
	}
	return &ChannelRx{
		ChannelRx:        *rrpitTransferChannel.NewChannelRx(channelID),
		OnNewDataMessage: onNewDataMessage,
	}, nil
}

func (rx *ChannelRx) OnNewMessageArrived(message []byte) (err error) {
	if rx.closed {
		return newError("channel receiver closed")
	}
	if len(message) < channelSequenceFieldLength {
		return newError("materialized channel message too short")
	}

	dataMessage := rrpitTransferChannel.ChannelDataMessage{
		ChannelSeq: binary.BigEndian.Uint64(message[:channelSequenceFieldLength]),
		Data:       append([]byte(nil), message[channelSequenceFieldLength:]...),
	}
	if err := rx.ProcessMessageReceived(dataMessage); err != nil {
		return err
	}
	return rx.OnNewDataMessage(dataMessage.Data)
}

func (rx *ChannelRx) AssignChannelID(channelID uint64) error {
	return rx.ChannelRx.AssignChannelID(channelID)
}

func (rx *ChannelRx) Close() error {
	rx.closed = true
	return nil
}

type ChannelTx struct {
	rrpitTransferChannel.ChannelTx
	io.WriteCloser
}

func NewChannelTx(channelID uint64, writer io.WriteCloser, maxRewindableTimestampNum int, maxRewindableControlMessageNum int) (*ChannelTx, error) {
	if writer == nil {
		return nil, newError("nil writer")
	}
	channelTx, err := rrpitTransferChannel.NewChannelTx(channelID, maxRewindableTimestampNum, maxRewindableControlMessageNum)
	if err != nil {
		return nil, err
	}
	return &ChannelTx{
		ChannelTx:   *channelTx,
		WriteCloser: writer,
	}, nil
}

func (tx *ChannelTx) SendDataMessage(data []byte) error {
	payload := append([]byte(nil), data...)
	wireMessage := make([]byte, channelSequenceFieldLength+len(payload))
	binary.BigEndian.PutUint64(wireMessage[:channelSequenceFieldLength], tx.NextSeq)
	copy(wireMessage[channelSequenceFieldLength:], payload)

	written, err := tx.Write(wireMessage)
	if err != nil {
		return err
	}
	if written != len(wireMessage) {
		return io.ErrShortWrite
	}
	if _, err := tx.CreateDataMessage(payload, uint64(time.Now().UnixNano())); err != nil {
		return err
	}
	return nil
}
