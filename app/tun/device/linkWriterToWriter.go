package device

import (
	"io"

	"github.com/v2fly/v2ray-core/v5/common/errors"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func NewLinkWriterToWriter(writer stack.LinkWriter) io.Writer {
	return &linkWriterToWriter{writer: writer}
}

type linkWriterToWriter struct {
	writer stack.LinkWriter
}

func (l linkWriterToWriter) Write(p []byte) (n int, err error) {
	buffer := buffer.MakeWithData(p)
	packetBufferPtr := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer,
		OnRelease: func() {
			buffer.Release()
		},
	})
	packetList := stack.PacketBufferList{}
	packetList.PushBack(packetBufferPtr)
	_, terr := l.writer.WritePackets(packetList)
	if terr != nil {
		return 0, newError("failed to write packet").Base(errors.New(terr.String())).AtError()
	}
	return len(p), nil
}
