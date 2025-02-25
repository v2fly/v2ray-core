package mirrorcommon

import (
	"bytes"

	"github.com/v2fly/struc"
)

type TLSClientHello struct {
	HandshakeType uint8
	Length        [3]byte
	Version       uint16
	ClientRandom  [32]byte

	// There are other entries, however we do not need them yet
}

func UnpackTLSClientHello(data []byte) (TLSClientHello, error) {
	var clientHello TLSClientHello
	err := struc.Unpack(bytes.NewReader(data), &clientHello)
	return clientHello, err
}

type TLSServerHello struct {
	HandshakeType   uint8
	Length          [3]byte
	Version         uint16
	ServerRandom    [32]byte
	SessionIDLength uint8 `struc:"sizeof=SessionID"`
	SessionID       []byte
	CipherSuite     uint16

	// There are other entries, however we do not need them yet
}

func UnpackTLSServerHello(data []byte) (TLSServerHello, error) {
	var serverHello TLSServerHello
	err := struc.Unpack(bytes.NewReader(data), &serverHello)
	return serverHello, err
}
