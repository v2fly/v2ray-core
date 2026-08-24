package udp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/transport"
	. "github.com/v2fly/v2ray-core/v5/transport/internet/udp"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

func TestPacketAddrDispatcherHandlesFailedDispatch(t *testing.T) {
	td := &TestDispatcher{
		OnDispatch: func(ctx context.Context, dest net.Destination) (*transport.Link, error) {
			return nil, errors.New("dispatch rejected")
		},
	}

	creator := NewPacketAddrDispatcherCreator(context.Background())
	dispatcher := creator.NewPacketAddrDispatcher(td, func(ctx context.Context, packet *udp.Packet) {
		t.Error("unexpected response callback for a failed dispatch")
	})
	if dispatcher == nil {
		t.Fatal("NewPacketAddrDispatcher returned nil")
	}

	b := buf.New()
	common.Must2(b.WriteString("abcd"))
	dispatcher.Dispatch(context.Background(), net.UDPDestination(net.LocalHostIP, 53), b)

	common.Must(dispatcher.Close())
	common.Must(dispatcher.Close())
}

func TestPacketAddrDispatcherReleasesPayload(t *testing.T) {
	uplinkReader, uplinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	downlinkReader, downlinkWriter := pipe.New(pipe.WithSizeLimit(1024))
	defer common.Interrupt(uplinkWriter)
	defer common.Interrupt(downlinkWriter)

	td := &TestDispatcher{
		OnDispatch: func(ctx context.Context, dest net.Destination) (*transport.Link, error) {
			return &transport.Link{Reader: downlinkReader, Writer: uplinkWriter}, nil
		},
	}

	creator := NewPacketAddrDispatcherCreator(context.Background())
	dispatcher := creator.NewPacketAddrDispatcher(td, func(ctx context.Context, packet *udp.Packet) {})
	defer dispatcher.Close()

	b := buf.New()
	common.Must2(b.WriteString("abcd"))
	dispatcher.Dispatch(context.Background(), net.UDPDestination(net.LocalHostIP, 53), b)
	if !b.IsEmpty() {
		t.Fatal("Dispatch did not take ownership of the payload buffer")
	}

	mb, err := uplinkReader.ReadMultiBuffer()
	common.Must(err)
	buf.ReleaseMulti(mb)
}
