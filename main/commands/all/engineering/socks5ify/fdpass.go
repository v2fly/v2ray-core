//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"

	"golang.org/x/sys/unix"
)

const (
	tunFDMessage     = "tun"
	childErrorPrefix = "error:"
	maxReceivedFDs   = 16
)

func sendFileDescriptor(sock int, fd int) error {
	rights := unix.UnixRights(fd)
	return unix.Sendmsg(sock, []byte(tunFDMessage), rights, nil, 0)
}

func sendChildError(sock int, err error) {
	_ = unix.Sendmsg(sock, []byte(childErrorPrefix+err.Error()), nil, nil, 0)
}

func recvFileDescriptor(sock int) (int, error) {
	payload := make([]byte, 256)
	oob := make([]byte, unix.CmsgSpace(4*maxReceivedFDs))
	n, oobn, flags, _, err := unix.Recvmsg(sock, payload, oob, 0)
	if err != nil {
		return -1, err
	}

	fds, err := parseFileDescriptors(oob[:oobn])
	if err != nil {
		return -1, err
	}
	if flags&unix.MSG_CTRUNC != 0 {
		closeFileDescriptors(fds)
		return -1, fmt.Errorf("file descriptor message truncated")
	}
	if flags&unix.MSG_TRUNC != 0 {
		closeFileDescriptors(fds)
		return -1, fmt.Errorf("child message truncated")
	}

	if n >= len(childErrorPrefix) && string(payload[:len(childErrorPrefix)]) == childErrorPrefix {
		closeFileDescriptors(fds)
		return -1, fmt.Errorf("%s", string(payload[len(childErrorPrefix):n]))
	}
	if string(payload[:n]) != tunFDMessage {
		closeFileDescriptors(fds)
		return -1, fmt.Errorf("unexpected child message %q", string(payload[:n]))
	}
	if len(fds) == 0 {
		return -1, fmt.Errorf("child did not send a file descriptor")
	}
	closeFileDescriptors(fds[1:])
	return fds[0], nil
}

func parseFileDescriptors(oob []byte) ([]int, error) {
	messages, err := unix.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, err
	}
	var allFDs []int
	for _, message := range messages {
		if message.Header.Level != unix.SOL_SOCKET || message.Header.Type != unix.SCM_RIGHTS {
			continue
		}
		fds, err := unix.ParseUnixRights(&message)
		if err != nil {
			closeFileDescriptors(allFDs)
			return nil, err
		}
		allFDs = append(allFDs, fds...)
	}
	return allFDs, nil
}

func closeFileDescriptors(fds []int) {
	for _, fd := range fds {
		_ = unix.Close(fd)
	}
}
