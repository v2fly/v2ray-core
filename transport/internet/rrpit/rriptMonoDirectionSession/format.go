package rriptMonoDirectionSession

const (
	PacketKind_UNDEFINED uint8 = 0

	PacketKind_InteractiveStreamData    uint8 = 1
	PacketKind_InteractiveStreamControl uint8 = 2
	PacketKind_BackgroundStreamData     uint8 = 3
	PacketKind_BackgroundStreamControl  uint8 = 4
	PacketKind_InteractivePacketData    uint8 = 5
	PacketKind_InteractivePacketControl uint8 = 6
	PacketKind_BackgroundPacketData     uint8 = 7
	PacketKind_BackgroundPacketControl  uint8 = 8

	PacketKind_DATA    uint8 = PacketKind_InteractiveStreamData
	PacketKind_CONTROL uint8 = PacketKind_InteractiveStreamControl
)

type SessionPacket struct {
	PacketKind uint8
	// followed by either a control packet content or data packet content
}

func PacketKindName(kind uint8) string {
	switch kind {
	case PacketKind_InteractiveStreamData:
		return "interactive_stream_data"
	case PacketKind_InteractiveStreamControl:
		return "interactive_stream_control"
	case PacketKind_BackgroundStreamData:
		return "background_stream_data"
	case PacketKind_BackgroundStreamControl:
		return "background_stream_control"
	case PacketKind_InteractivePacketData:
		return "interactive_packet_data"
	case PacketKind_InteractivePacketControl:
		return "interactive_packet_control"
	case PacketKind_BackgroundPacketData:
		return "background_packet_data"
	case PacketKind_BackgroundPacketControl:
		return "background_packet_control"
	default:
		return "unknown"
	}
}

func IsControlPacketKind(kind uint8) bool {
	switch kind {
	case PacketKind_InteractiveStreamControl,
		PacketKind_BackgroundStreamControl,
		PacketKind_InteractivePacketControl,
		PacketKind_BackgroundPacketControl:
		return true
	default:
		return false
	}
}

func IsDataPacketKind(kind uint8) bool {
	switch kind {
	case PacketKind_InteractiveStreamData,
		PacketKind_BackgroundStreamData,
		PacketKind_InteractivePacketData,
		PacketKind_BackgroundPacketData:
		return true
	default:
		return false
	}
}
