package webrtc

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"time"

	pionice "github.com/pion/ice/v4"
	pionwebrtc "github.com/pion/webrtc/v4"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

const (
	portBlossomFirstPort = 1
	portBlossomLastPort  = 65535
)

var portBlossomRepeatInterval = time.Second

func candidateBlossomIP(raw []byte) (net.IP, error) {
	var candidateInit pionwebrtc.ICECandidateInit
	if err := json.Unmarshal(raw, &candidateInit); err != nil {
		return nil, newError("failed to decode candidate for port blossom").Base(err)
	}
	if candidateInit.Candidate == "" {
		return nil, newError("candidate for port blossom is empty")
	}

	candidate, err := pionice.UnmarshalCandidate(candidateInit.Candidate)
	if err != nil {
		return nil, newError("failed to parse candidate for port blossom").Base(err)
	}

	ip := net.ParseIP(candidate.Address())
	if ip == nil {
		return nil, newError("candidate address is not an IP: ", candidate.Address())
	}
	if ip.IsUnspecified() {
		return nil, newError("candidate address is unspecified")
	}

	return ip, nil
}

func blossomUDPPorts(packetConns []v2net.PacketConn, ip net.IP) error {
	if ip == nil {
		return newError("missing candidate IP for port blossom")
	}

	sent := false
	var firstErr error
	for _, packetConn := range packetConns {
		if packetConn == nil || !packetConnSupportsIP(packetConn, ip) {
			continue
		}
		sent = true
		for port := portBlossomFirstPort; port <= portBlossomLastPort; port++ {
			payload := portBlossomPayload(port)
			if _, err := packetConn.WriteTo(payload, &v2net.UDPAddr{
				IP:   ip,
				Port: port,
			}); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				break
			}
		}
	}

	if !sent {
		return newError("no listener packet socket available for candidate IP ", ip.String())
	}

	return firstErr
}

func portBlossomPayload(port int) []byte {
	var payload [2]byte
	binary.BigEndian.PutUint16(payload[:], uint16(port))
	return payload[:]
}

func packetConnSupportsIP(packetConn v2net.PacketConn, ip net.IP) bool {
	udpAddr, ok := packetConn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr == nil || udpAddr.IP == nil {
		return true
	}

	if udpAddr.IP.To4() != nil {
		return ip.To4() != nil
	}

	if udpAddr.IP.To16() != nil {
		return ip.To4() == nil
	}

	return true
}
