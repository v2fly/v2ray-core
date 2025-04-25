//go:build freebsd
// +build freebsd

package udp

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/ghxhy/v2ray-core/v5/common/errors"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

// RetrieveOriginalDest from stored laddr, caddr
func RetrieveOriginalDest(oob []byte) net.Destination {
	dec := gob.NewDecoder(bytes.NewBuffer(oob))
	var la, ra net.UDPAddr
	dec.Decode(&la)
	dec.Decode(&ra)
	ip, port, err := internet.OriginalDst(&la, &ra)
	if err != nil {
		return net.Destination{}
	}
	return net.UDPDestination(net.IPAddress(ip), net.Port(port))
}

// ReadUDPMsg stores laddr, caddr for later use
func ReadUDPMsg(conn *net.UDPConn, payload []byte, oob []byte) (int, int, int, *net.UDPAddr, error) {
	nBytes, addr, err := conn.ReadFromUDP(payload)
	if err != nil {
		return nBytes, 0, 0, addr, err
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return 0, 0, 0, nil, errors.New("invalid local address")
	}
	if addr == nil {
		return 0, 0, 0, nil, errors.New("invalid remote address")
	}
	enc.Encode(localAddr)
	enc.Encode(addr)

	var reader io.Reader = &buf
	noob, _ := reader.Read(oob)

	return nBytes, noob, 0, addr, nil
}
