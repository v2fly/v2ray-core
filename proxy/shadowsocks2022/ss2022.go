package shadowsocks2022

import (
	"crypto/cipher"
	"github.com/lunixbochs/struc"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"io"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type KeyDerivation interface {
	GetSessionSubKey(effectivePsk, Salt []byte, OutKey []byte) error
	GetIdentitySubKey(effectivePsk, Salt []byte, OutKey []byte) error
}

type Method interface {
	GetSessionSubKeyAndSaltLength() int
	GetStreamAEAD(SessionSubKey []byte) (cipher.AEAD, error)
	GenerateEIH(CurrentIdentitySubKey []byte, nextPskHash []byte, out []byte) error
	GetUDPClientProcessor(ipsk [][]byte, psk []byte, derivation KeyDerivation) (UDPClientPacketProcessor, error)
}

type ExtensibleIdentityHeaders interface {
	struc.Custom
}

type DestinationAddress interface {
	net.Address
}

type RequestSalt interface {
	struc.Custom
	isRequestSalt()
	Bytes() []byte
	FillAllFrom(reader io.Reader) error
}

type TCPRequestHeader1PreSessionKey struct {
	Salt RequestSalt
	EIH  ExtensibleIdentityHeaders
}

type TCPRequestHeader2FixedLength struct {
	Type         byte
	Timestamp    uint64
	HeaderLength uint16
}

type TCPRequestHeader3VariableLength struct {
	DestinationAddress DestinationAddress
	Contents           struct {
		PaddingLength uint16 `struc:"sizeof=Padding"`
		Padding       []byte
	}
}

type TCPRequestHeader struct {
	PreSessionKeyHeader TCPRequestHeader1PreSessionKey
	FixedLengthHeader   TCPRequestHeader2FixedLength
	Header              TCPRequestHeader3VariableLength
}

type TCPResponseHeader1PreSessionKey struct {
	Salt RequestSalt
}

type TCPResponseHeader2FixedLength struct {
	Type                 byte
	Timestamp            uint64
	RequestSalt          RequestSalt
	InitialPayloadLength uint16
}
type TCPResponseHeader struct {
	PreSessionKeyHeader TCPResponseHeader1PreSessionKey
	Header              TCPResponseHeader2FixedLength
}

const TCPHeaderTypeClientToServerStream = byte(0x00)
const TCPHeaderTypeServerToClientStream = byte(0x01)
const TCPMinPaddingLength = 0
const TCPMaxPaddingLength = 900

var addrParser = protocol.NewAddressParser(
	protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(0x04, net.AddressFamilyIPv6),
	protocol.AddressFamilyByte(0x03, net.AddressFamilyDomain),
)

type UDPRequest struct {
	SessionID [8]byte
	PacketID  uint64
	TimeStamp uint64
	Address   DestinationAddress
	Port      int
	Payload   *buf.Buffer
}

type UDPResponse struct {
	UDPRequest
	ClientSessionID [8]byte
}

const UDPHeaderTypeClientToServerStream = byte(0x00)
const UDPHeaderTypeServerToClientStream = byte(0x01)

// UDPClientPacketProcessor
// Caller retain and receive all ownership of the buffer
type UDPClientPacketProcessor interface {
	EncodeUDPRequest(request *UDPRequest, out *buf.Buffer) error
	DecodeUDPResp(input []byte, resp *UDPResponse) error
}
