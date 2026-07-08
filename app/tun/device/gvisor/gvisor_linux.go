//go:build linux && ((linux && amd64) || (linux && arm64))
// +build linux
// +build linux,amd64 linux,arm64

package gvisor

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/app/tun/device"

	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
)

const (
	ifReqSize = unix.IFNAMSIZ + 64
)

type GvisorTUN struct {
	stack.LinkEndpoint

	options device.Options

	closeAccess sync.Mutex
	fd          int
	mtu         uint32 // real MTU
}

func New(options device.Options) (device.Device, error) {
	t := &GvisorTUN{options: options, fd: -1}

	fd := -1
	preopened := options.PreopenedFDSet
	if preopened {
		if options.PreopenedFD < 0 {
			return nil, newError("invalid preopened tun file descriptor").AtError()
		}
		fd = options.PreopenedFD
	}

	if len(options.Name) > unix.IFNAMSIZ {
		if preopened {
			_ = unix.Close(fd)
		}
		return nil, newError("name too long").AtError()
	}

	if preopened {
		if err := unix.SetNonblock(fd, true); err != nil {
			_ = unix.Close(fd)
			return nil, newError("failed to set tun device non-blocking").Base(err).AtError()
		}
	} else {
		var err error
		fd, err = tun.Open(options.Name)
		if err != nil {
			return nil, newError("failed to open tun device").Base(err).AtError()
		}

		if options.MTU > 0 {
			_ = setMTU(options.Name, int(options.MTU))
		}
	}
	t.fd = fd

	mtu := options.MTU
	if !preopened {
		var err error
		mtu, err = rawfile.GetMTU(options.Name)
		if err != nil {
			_ = unix.Close(fd)
			return nil, newError("failed to get mtu").Base(err).AtError()
		}
	}
	if mtu == 0 {
		mtu = 1500
	}
	t.mtu = mtu

	linkEndpoint, err := fdbased.New(&fdbased.Options{
		FDs: []int{fd},
		MTU: mtu,
		// TUN is not need to process ethernet header.
		EthernetHeader: false,
		// Readv is the default dispatch mode and is the least performant of the
		// dispatch options but the one that is supported by all underlying FD
		// types.
		PacketDispatchMode:    fdbased.Readv,
		MaxSyscallHeaderBytes: 0x00,
	})
	if err != nil {
		_ = unix.Close(fd)
		return nil, newError("failed to create link endpoint").Base(err).AtError()
	}
	t.LinkEndpoint = linkEndpoint

	return t, nil
}

func (t *GvisorTUN) Close() {
	t.closeAccess.Lock()
	defer t.closeAccess.Unlock()
	if t.fd >= 0 {
		_ = unix.Close(t.fd)
		t.fd = -1
	}
}

// Modified from golang.zx2c4.com/wireguard/tun/tun_linux.go
func setMTU(name string, n int) error {
	// open datagram socket
	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_DGRAM|unix.SOCK_CLOEXEC,
		0,
	)
	if err != nil {
		return err
	}

	defer func() { _ = unix.Close(fd) }()

	// do ioctl call
	var ifr [ifReqSize]byte
	copy(ifr[:], name)
	*(*uint32)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = uint32(n)
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFMTU),
		uintptr(unsafe.Pointer(&ifr[0])),
	)

	if errno != 0 {
		return fmt.Errorf("failed to set MTU of TUN device: %w", errno)
	}

	return nil
}
