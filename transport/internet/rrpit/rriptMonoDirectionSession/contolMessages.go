package rriptMonoDirectionSession

import (
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferChannel"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitTransferLane"
)

type SessionInstanceID [16]byte

type ControlMessage struct {
	Session      SessionControlMessage
	FloodChannel SessionFloodChannelControlMessage
	Lane         SessionLaneControlMessage
	Channel      SessionChannelControlMessage
}

type SessionControlMessage struct {
	InstanceID SessionInstanceID
}

type SessionFloodChannelControlMessage struct {
	CurrentChannelID uint64
}

type SessionLaneControlMessage struct {
	LaneACKTo      int64
	LenLaneControl uint16 `struc:"uint16,sizeof=LaneControl"`
	LaneControl    []rrpitTransferLane.TransferControl
}

type SessionChannelControlMessage struct {
	LenChannelControl uint16 `struc:"uint16,sizeof=ChannelControl"`
	ChannelControl    []rrpitTransferChannel.ChannelControlMessage
}
