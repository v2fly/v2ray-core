package rriptMonoDirectionSession

const (
	PacketKind_UNDEFINED uint8 = 0
	PacketKind_DATA      uint8 = 1
	PacketKind_CONTROL   uint8 = 2
)

type SessionPacket struct {
	PacketKind uint8
	// followed by either a control packet content or data packet content
}
