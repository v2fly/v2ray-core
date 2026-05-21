package realm

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
)

func NewRealmPeer(scheme, host, port, token, id string, stunServers []string, raw net.PacketConn) (*net.UDPAddr, error) {
	start := time.Now()
	servers := resolveSTUNServers(raw.LocalAddr().(*net.UDPAddr).IP, stunServers)
	newError("[realm] get stun servers ", servers, " with ", time.Since(start)).WriteToLog()
	if len(servers) == 0 {
		return nil, errors.New("empty stun servers")
	}

	start = time.Now()
	locals := Discover(raw, servers)
	newError("[realm] get stun locals ", locals, " with ", time.Since(start)).WriteToLog()
	if len(locals) == 0 {
		return nil, errors.New("empty stun locals")
	}

	rClient, err := NewClient(scheme, host, port, token)
	if err != nil {
		return nil, newError("http create").Base(err)
	}

	meta := common.Must2(NewPunchMetadata()).(PunchMetadata)

	start = time.Now()
	resp, err := rClient.Connect(context.Background(), id, ConnectRequest{
		Addresses:     addrPortStrings(locals),
		PunchMetadata: meta,
	})
	if err != nil {
		return nil, newError("http connect").Base(err)
	}
	newError("[realm] ", id, " ", meta.Nonce, " connect ", resp.Addresses, " with ", time.Since(start)).WriteToLog()

	peers, _ := parseAddrPorts(resp.Addresses)
	newError("[realm] get peers ", peers).AtDebug().WriteToLog()
	filteredPeers, seen := candidatePunchAddrs(locals, peers)
	newError("[realm] filtered peers ", filteredPeers).AtDebug().WriteToLog()
	expandedPeers := expandSymmetricNATCandidates(filteredPeers, seen)
	newError("[realm] expanded peers ", expandedPeers).AtDebug().WriteToLog()

	if len(expandedPeers) == 0 {
		return nil, errors.New("empty peers")
	}

	start = time.Now()
	result, err := Punch(raw, expandedPeers, meta, defaultPunchTimeout, defaultPunchInterval)
	if err != nil {
		return nil, newError("punch fail").Base(err)
	}
	newError("[realm] punch peer ", result, " with ", time.Since(start)).WriteToLog()

	return result, nil
}

func Punch(conn net.PacketConn, peers []netip.AddrPort, meta PunchMetadata, timeout, interval time.Duration) (*net.UDPAddr, error) {
	defer conn.SetReadDeadline(time.Time{})
	nextSend := time.Now()
	deadline := nextSend.Add(timeout)
	buf := make([]byte, punchMaxWireLen)
	for {
		now := time.Now()
		if now.After(deadline) {
			return nil, errors.New("timeout")
		}
		if now.After(nextSend) {
			sendPunchPackets(conn, peers, meta, PunchPacketHello)
			nextSend = now.Add(interval)
		}

		if nextSend.After(deadline) {
			conn.SetReadDeadline(deadline)
		} else {
			conn.SetReadDeadline(nextSend)
		}
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			return nil, err
		}
		packet, err := DecodePunchPacket(buf[:n], meta)
		if err != nil {
			continue
		}
		if packet.Type == PunchPacketHello {
			sendPunchPacket(conn, addr.(*net.UDPAddr).AddrPort(), meta, PunchPacketAck)
		}
		return addr.(*net.UDPAddr), nil
	}
}
