package internet

import (
	"syscall"
)

const (
	// TCP_FASTOPEN is the socket option on darwin for TCP fast open.
	TCP_FASTOPEN = 0x105 // nolint: golint,stylecheck
	// TCP_FASTOPEN_SERVER is the value to enable TCP fast open on darwin for server connections.
	TCP_FASTOPEN_SERVER = 0x01 // nolint: golint,stylecheck
	// TCP_FASTOPEN_CLIENT is the value to enable TCP fast open on darwin for client connections.
	TCP_FASTOPEN_CLIENT = 0x02  // nolint: golint,stylecheck
	TCP_KEEPINTVL       = 0x101 // nolint: golint,stylecheck
)

func applyOutboundSocketOptions(network string, address string, fd uintptr, config *SocketConfig) error {
	if isTCPSocket(network) {
		switch config.Tfo {
		case SocketConfig_Enable:
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_FASTOPEN, TCP_FASTOPEN_CLIENT); err != nil {
				return err
			}
		case SocketConfig_Disable:
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_FASTOPEN, 0); err != nil {
				return err
			}
		}

		if config.TcpKeepAliveInterval > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return newError("failed to set TCP_KEEPINTVL", err)
			}
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
				return newError("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	return nil
}

func applyInboundSocketOptions(network string, fd uintptr, config *SocketConfig) error {
	if isTCPSocket(network) {
		switch config.Tfo {
		case SocketConfig_Enable:
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_FASTOPEN, TCP_FASTOPEN_SERVER); err != nil {
				return err
			}
		case SocketConfig_Disable:
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_FASTOPEN, 0); err != nil {
				return err
			}
		}
		if config.TcpKeepAliveInterval > 0 {
			if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_KEEPINTVL, int(config.TcpKeepAliveInterval)); err != nil {
				return newError("failed to set TCP_KEEPINTVL", err)
			}
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
				return newError("failed to set SO_KEEPALIVE", err)
			}
		}
	}

	return nil
}

func bindAddr(fd uintptr, address []byte, port uint32) error {
	return nil
}

func setReuseAddr(fd uintptr) error {
	return nil
}

func setReusePort(fd uintptr) error {
	return nil
}
