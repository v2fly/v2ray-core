package packetaddr

import (
	sysnet "net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/v2fly/v2ray-core/v5/common/buf"
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
