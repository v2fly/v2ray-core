package webrtc

import (
	"encoding/binary"
)

const (
	frameTypeOpen  = 1
	frameTypeData  = 2
	frameTypeClose = 3
)

type wireFrame struct {
	FrameType byte
	StreamID  uint32
	Tag       string
	Payload   []byte
}

func encodeWireFrame(frame wireFrame) []byte {
	switch frame.FrameType {
	case frameTypeOpen:
		tagBytes := []byte(frame.Tag)
		out := make([]byte, 1+4+2+len(tagBytes))
		out[0] = frame.FrameType
		binary.BigEndian.PutUint32(out[1:], frame.StreamID)
		binary.BigEndian.PutUint16(out[5:], uint16(len(tagBytes)))
		copy(out[7:], tagBytes)
		return out
	case frameTypeData:
		out := make([]byte, 1+4+4+len(frame.Payload))
		out[0] = frame.FrameType
		binary.BigEndian.PutUint32(out[1:], frame.StreamID)
		binary.BigEndian.PutUint32(out[5:], uint32(len(frame.Payload)))
		copy(out[9:], frame.Payload)
		return out
	case frameTypeClose:
		out := make([]byte, 1+4)
		out[0] = frame.FrameType
		binary.BigEndian.PutUint32(out[1:], frame.StreamID)
		return out
	default:
		return nil
	}
}

func decodeWireFrame(data []byte) (wireFrame, error) {
	if len(data) < 5 {
		return wireFrame{}, newError("short wire frame")
	}

	frame := wireFrame{
		FrameType: data[0],
		StreamID:  binary.BigEndian.Uint32(data[1:5]),
	}

	switch frame.FrameType {
	case frameTypeOpen:
		if len(data) < 7 {
			return wireFrame{}, newError("short open frame")
		}
		tagLen := int(binary.BigEndian.Uint16(data[5:7]))
		if len(data) != 7+tagLen {
			return wireFrame{}, newError("invalid open frame length")
		}
		frame.Tag = string(data[7:])
		return frame, nil
	case frameTypeData:
		if len(data) < 9 {
			return wireFrame{}, newError("short data frame")
		}
		payloadLen := int(binary.BigEndian.Uint32(data[5:9]))
		if len(data) != 9+payloadLen {
			return wireFrame{}, newError("invalid data frame length")
		}
		frame.Payload = data[9:]
		return frame, nil
	case frameTypeClose:
		if len(data) != 5 {
			return wireFrame{}, newError("invalid close frame length")
		}
		return frame, nil
	default:
		return wireFrame{}, newError("unknown frame type ", frame.FrameType)
	}
}
