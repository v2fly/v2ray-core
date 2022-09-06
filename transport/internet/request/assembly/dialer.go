package assembly

import (
	"context"
	gonet "net"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

type client struct {
	tripper   request.RoundTripperClient
	assembler request.SessionAssemblerClient

	streamSettings *internet.MemoryStreamConfig
	dest           net.Destination
}

func (c client) Dial(ctx context.Context) (net.Conn, error) {
	return transportcommon.DialWithSecuritySettings(ctx, c.dest, c.streamSettings)
}

func (c client) AutoImplDialer() request.Dialer {
	return c
}

func (c client) Tripper() request.Tripper {
	return c.tripper
}

func (c client) dialRequestSession(ctx context.Context) (net.Conn, error) {
	session, err := c.assembler.NewSession(ctx)
	if err != nil {
		return nil, newError("failed to create new session").Base(err)
	}
	return clientConnection{session}, nil
}

type clientConnection struct {
	request.Session
}

func (c clientConnection) LocalAddr() gonet.Addr {
	return &net.UnixAddr{Name: "unimplemented"}
}

func (c clientConnection) RemoteAddr() gonet.Addr {
	return &net.UnixAddr{Name: "unimplemented"}
}

func (c clientConnection) SetDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func (c clientConnection) SetReadDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func (c clientConnection) SetWriteDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func dialRequest(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	clientAssembly := &client{}
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)

	assemblerConfigInstance, err := serial.GetInstanceOf(transportConfiguration.Assembler)
	if err != nil {
		return nil, newError("failed to get config instance of assembler").Base(err)
	}
	assembler, err := common.CreateObject(ctx, assemblerConfigInstance)
	if err != nil {
		return nil, newError("failed to create assembler").Base(err)
	}
	if typedAssembler, ok := assembler.(request.SessionAssemblerClient); !ok {
		return nil, newError("failed to type assert assembler to SessionAssemblerClient")
	} else {
		clientAssembly.assembler = typedAssembler
	}

	roundtripperConfigInstance, err := serial.GetInstanceOf(transportConfiguration.Roundtripper)
	if err != nil {
		return nil, newError("failed to get config instance of roundtripper").Base(err)
	}
	roundtripper, err := common.CreateObject(ctx, roundtripperConfigInstance)
	if err != nil {
		return nil, newError("failed to create roundtripper").Base(err)
	}
	if typedRoundtripper, ok := roundtripper.(request.RoundTripperClient); !ok {
		return nil, newError("failed to type assert roundtripper to RoundTripperClient")
	} else {
		clientAssembly.tripper = typedRoundtripper
	}

	clientAssembly.streamSettings = streamSettings
	clientAssembly.dest = dest

	clientAssembly.assembler.OnTransportClientAssemblyReady(clientAssembly)
	clientAssembly.tripper.OnTransportClientAssemblyReady(clientAssembly)
	return clientAssembly.dialRequestSession(ctx)
}

func dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialRequest(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial request to ", dest).Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, dial))
}
