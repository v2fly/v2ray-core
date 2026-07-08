//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"bytes"
	"encoding/binary"
	"fmt"
	stdnet "net"
	"syscall"

	"golang.org/x/sys/unix"
)

func interfaceIndex(name string) (int, error) {
	iface, err := stdnet.InterfaceByName(name)
	if err != nil {
		return 0, err
	}
	return iface.Index, nil
}

func addAddress(ifindex int, ipText string, prefix int) error {
	ip := stdnet.ParseIP(ipText)
	if ip == nil {
		return fmt.Errorf("invalid IP address %q", ipText)
	}

	family := uint8(unix.AF_INET)
	addr := ip.To4()
	if addr == nil {
		family = unix.AF_INET6
		addr = ip.To16()
	}
	if addr == nil {
		return fmt.Errorf("invalid IP address %q", ipText)
	}

	msg := unix.IfAddrmsg{
		Family:    family,
		Prefixlen: uint8(prefix),
		Scope:     unix.RT_SCOPE_UNIVERSE,
		Index:     uint32(ifindex),
	}
	payload := append(marshal(msg),
		rtAttr(unix.IFA_LOCAL, addr)...)
	payload = append(payload, rtAttr(unix.IFA_ADDRESS, addr)...)
	return netlinkRequest(unix.RTM_NEWADDR, unix.NLM_F_REQUEST|unix.NLM_F_ACK|unix.NLM_F_CREATE|unix.NLM_F_REPLACE, payload)
}

func addDefaultRoute(ifindex int, ipv6 bool) error {
	family := uint8(unix.AF_INET)
	if ipv6 {
		family = unix.AF_INET6
	}
	msg := unix.RtMsg{
		Family:   family,
		Table:    unix.RT_TABLE_MAIN,
		Protocol: unix.RTPROT_BOOT,
		Scope:    unix.RT_SCOPE_LINK,
		Type:     unix.RTN_UNICAST,
	}
	oif := make([]byte, 4)
	binary.NativeEndian.PutUint32(oif, uint32(ifindex))
	payload := append(marshal(msg), rtAttr(unix.RTA_OIF, oif)...)
	return netlinkRequest(unix.RTM_NEWROUTE, unix.NLM_F_REQUEST|unix.NLM_F_ACK|unix.NLM_F_CREATE|unix.NLM_F_REPLACE, payload)
}

func netlinkRequest(messageType uint16, flags int, payload []byte) error {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_ROUTE)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	seq := uint32(1)
	header := unix.NlMsghdr{
		Len:   uint32(unix.NLMSG_HDRLEN + len(payload)),
		Type:  messageType,
		Flags: uint16(flags),
		Seq:   seq,
		Pid:   uint32(unix.Getpid()),
	}
	request := append(marshal(header), payload...)
	if err := unix.Sendto(fd, request, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return err
	}

	response := make([]byte, 8192)
	for {
		n, _, err := unix.Recvfrom(fd, response, 0)
		if err != nil {
			return err
		}
		messages, err := syscall.ParseNetlinkMessage(response[:n])
		if err != nil {
			return err
		}
		for _, message := range messages {
			if message.Header.Seq != seq {
				continue
			}
			switch message.Header.Type {
			case unix.NLMSG_ERROR:
				var errMessage unix.NlMsgerr
				if err := binary.Read(bytes.NewReader(message.Data), binary.NativeEndian, &errMessage); err != nil {
					return err
				}
				if errMessage.Error == 0 {
					return nil
				}
				return syscall.Errno(-errMessage.Error)
			case unix.NLMSG_DONE:
				return nil
			}
		}
	}
}

func rtAttr(attrType int, data []byte) []byte {
	length := unix.SizeofRtAttr + len(data)
	attr := unix.RtAttr{
		Len:  uint16(length),
		Type: uint16(attrType),
	}
	out := append(marshal(attr), data...)
	for len(out)%4 != 0 {
		out = append(out, 0)
	}
	return out
}

func marshal(value interface{}) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.NativeEndian, value)
	return buf.Bytes()
}
