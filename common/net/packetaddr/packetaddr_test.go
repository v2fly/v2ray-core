package packetaddr

import (
	"bytes"
	"encoding/binary"
	"io"
	sysnet "net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

func TestPacketEncodingIPv4(t *testing.T) {
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv4(1, 2, 3, 4).To4(),
		Port: 1234,
	}
	var packetData [256]byte
	wrapped, err := AttachAddressToPacket(buf.FromBytes(packetData[:]), packetAddress)
	assert.NoError(t, err)

	packetPayload, decodedAddress, err := ExtractAddressFromPacket(wrapped)
	assert.NoError(t, err)

	assert.Equal(t, packetPayload.Bytes(), packetData[:])
	assert.Equal(t, packetAddress, decodedAddress)
}

func TestPacketEncodingIPv6(t *testing.T) {
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv6linklocalallrouters,
		Port: 1234,
	}
	var packetData [256]byte
	wrapped, err := AttachAddressToPacket(buf.FromBytes(packetData[:]), packetAddress)
	assert.NoError(t, err)

	packetPayload, decodedAddress, err := ExtractAddressFromPacket(wrapped)
	assert.NoError(t, err)

	assert.Equal(t, packetPayload.Bytes(), packetData[:])
	assert.Equal(t, packetAddress, decodedAddress)
}

func TestStreamPacketEncodingFragmented(t *testing.T) {
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv4(1, 2, 3, 4).To4(),
		Port: 1234,
	}
	packetData := []byte("hello over streamaddr")
	wrapped, err := AttachAddressToPacket(buf.FromBytes(packetData), packetAddress)
	assert.NoError(t, err)
	framed, err := AttachLengthToPacket(wrapped)
	assert.NoError(t, err)
	frameBytes := append([]byte(nil), framed.Bytes()...)
	framed.Release()

	reader := io.MultiReader(
		bytes.NewReader(frameBytes[:1]),
		bytes.NewReader(frameBytes[1:3]),
		bytes.NewReader(frameBytes[3:]),
	)
	packetPayload, err := ExtractPacketFromStream(reader)
	assert.NoError(t, err)

	decodedPayload, decodedAddress, err := ExtractAddressFromPacket(packetPayload)
	assert.NoError(t, err)

	assert.Equal(t, packetData, decodedPayload.Bytes())
	assert.Equal(t, packetAddress, decodedAddress)
}

func TestStreamPacketEncodingRejectsEmptyFrame(t *testing.T) {
	_, err := ExtractPacketFromStream(bytes.NewReader([]byte{0, 0}))
	assert.Error(t, err)
}

func TestStreamPacketConnWrapperWriteFragmented(t *testing.T) {
	conn := newTestPacketConn()
	wrapper := &streamPacketConnWrapper{PacketConn: conn}
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv4(1, 2, 3, 4).To4(),
		Port: 1234,
	}
	packetData := []byte("fragmented write")
	wrapped, err := AttachAddressToPacket(buf.FromBytes(packetData), packetAddress)
	assert.NoError(t, err)
	framed, err := AttachLengthToPacket(wrapped)
	assert.NoError(t, err)
	frameBytes := append([]byte(nil), framed.Bytes()...)
	framed.Release()

	n, err := wrapper.Write(frameBytes[:3])
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	select {
	case <-conn.writeCh:
		t.Fatal("unexpected UDP write before a complete stream frame")
	default:
	}

	n, err = wrapper.Write(frameBytes[3:])
	assert.NoError(t, err)
	assert.Equal(t, len(frameBytes)-3, n)
	written := <-conn.writeCh
	assert.Equal(t, packetData, written.payload)
	assert.Equal(t, packetAddress, written.addr)
}

func TestStreamPacketConnWrapperReadPartial(t *testing.T) {
	conn := newTestPacketConn()
	wrapper := &streamPacketConnWrapper{PacketConn: conn}
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv4(1, 2, 3, 4).To4(),
		Port: 1234,
	}
	packetData := []byte("partial read")
	conn.readCh <- testPacket{payload: packetData, addr: packetAddress}

	var streamData []byte
	readBuf := make([]byte, 3)
	for len(streamData) < streamPacketLengthSize || len(streamData) < streamPacketLengthSize+int(binary.BigEndian.Uint16(streamData[:streamPacketLengthSize])) {
		n, err := wrapper.Read(readBuf)
		assert.NoError(t, err)
		streamData = append(streamData, readBuf[:n]...)
	}

	packetPayload, err := ExtractPacketFromStream(bytes.NewReader(streamData))
	assert.NoError(t, err)
	decodedPayload, decodedAddress, err := ExtractAddressFromPacket(packetPayload)
	assert.NoError(t, err)
	assert.Equal(t, packetData, decodedPayload.Bytes())
	assert.Equal(t, packetAddress, decodedAddress)
}

func TestGetDestinationSubsetOfStream(t *testing.T) {
	isStream, err := GetDestinationSubsetOf(v2net.Destination{
		Network: v2net.Network_TCP,
		Address: v2net.DomainAddress(streamPacketMagicAddress),
		Port:    0,
	})
	assert.NoError(t, err)
	assert.True(t, isStream)
}

type testPacket struct {
	payload []byte
	addr    sysnet.Addr
}

type testPacketConn struct {
	readCh  chan testPacket
	writeCh chan testPacket
}

func newTestPacketConn() *testPacketConn {
	return &testPacketConn{
		readCh:  make(chan testPacket, 1),
		writeCh: make(chan testPacket, 1),
	}
}

func (c *testPacketConn) ReadFrom(p []byte) (int, sysnet.Addr, error) {
	packet, ok := <-c.readCh
	if !ok {
		return 0, nil, io.EOF
	}
	return copy(p, packet.payload), packet.addr, nil
}

func (c *testPacketConn) WriteTo(p []byte, addr sysnet.Addr) (int, error) {
	copied := append([]byte(nil), p...)
	c.writeCh <- testPacket{payload: copied, addr: addr}
	return len(p), nil
}

func (c *testPacketConn) Close() error {
	close(c.readCh)
	return nil
}

func (c *testPacketConn) LocalAddr() sysnet.Addr {
	return &sysnet.UDPAddr{}
}

func (c *testPacketConn) SetDeadline(time.Time) error {
	return nil
}

func (c *testPacketConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *testPacketConn) SetWriteDeadline(time.Time) error {
	return nil
}
