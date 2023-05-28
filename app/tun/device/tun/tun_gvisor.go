//go:build linux && ((linux && amd64) || (linux && arm64))
// +build linux
// +build linux,amd64 linux,arm64

package tun

import (
	"github.com/v2fly/v2ray-core/v5/app/tun/device"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
)

type TUN struct {
	stack.LinkEndpoint

	options device.Options

	fd  int
	mtu uint32 // real MTU
}

func New(options device.Options) (device.Device, error) {
	t := &TUN{options: options}

	if len(options.Name) > unix.IFNAMSIZ {
		return nil, newError("name too long").AtError()
	}

	fd, err := tun.Open(options.Name)
	if err != nil {
		return nil, newError("failed to open tun device").Base(err).AtError()
	}
	t.fd = fd

	// TODO: set MTU

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

func (t *TUN) Close() error {
	return unix.Close(t.fd)
}
