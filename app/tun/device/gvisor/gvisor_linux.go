//go:build linux && ((linux && amd64) || (linux && arm64))
// +build linux
// +build linux,amd64 linux,arm64

package gvisor

import (
	"fmt"
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

	fd  int
	mtu uint32 // real MTU
}

func New(options device.Options) (device.Device, error) {
	t := &GvisorTUN{options: options}

	if len(options.Name) > unix.IFNAMSIZ {
		return nil, newError("name too long").AtError()
	}

	fd, err := tun.Open(options.Name)
	if err != nil {
		return nil, newError("failed to open tun device").Base(err).AtError()
	}
	t.fd = fd

	if options.MTU > 0 {
		_ = setMTU(options.Name, int(options.MTU))
	}

	mtu, err := rawfile.GetMTU(options.Name)
	if err != nil {
		return nil, newError("failed to get mtu").Base(err).AtError()
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
		return nil, newError("failed to create link endpoint").Base(err).AtError()
	}
	t.LinkEndpoint = linkEndpoint

	return t, nil
}

func (t *GvisorTUN) Close() {
	_ = unix.Close(t.fd)
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
