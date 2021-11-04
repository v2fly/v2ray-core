package packetaddr

import (
	"github.com/stretchr/testify/assert"
	sysnet "net"
	"testing"
)

func TestPacketEncodingIPv4(t *testing.T) {
	packetAddress := &sysnet.UDPAddr{
		IP:   sysnet.IPv4(1, 2, 3, 4).To4(),
		Port: 1234,
	}
	var packetData [256]byte
	wrapped := AttachAddressToPacket(packetData[:], packetAddress)

	packetPayload, decodedAddress := ExtractAddressFromPacket(wrapped)

	assert.Equal(t, packetPayload, packetData[:])
	assert.Equal(t, packetAddress, decodedAddress)
}
