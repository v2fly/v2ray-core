//go:build js || dragonfly || netbsd || openbsd || solaris
// +build js dragonfly netbsd openbsd solaris

package internet

func applyOutboundSocketOptions(_ string, _ string, _ uintptr, _ *SocketConfig) error {
	return nil
}

func applyInboundSocketOptions(_ string, _ uintptr, _ *SocketConfig) error {
	return nil
}

func bindAddr(_ uintptr, _ []byte, _ uint32) error {
	return nil
}

func setReuseAddr(_ uintptr) error {
	return nil
}

func setReusePort(_ uintptr) error {
	return nil
}

func enableKeepAlive(_ uintptr, _ int32) error {
	return nil
}
