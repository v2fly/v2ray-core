package rrpitTransferChannel

type ChannelDataMessage struct {
	ChannelSeq uint64
	Data       []byte
}

type ChannelControlMessage struct {
	ChannelID                  uint64
	TotalPacketReceived        uint64
	LastSequenceNumberReceived uint64
}
