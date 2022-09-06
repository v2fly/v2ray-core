package assembly

import (
	"context"
	gonet "net"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

type server struct {
	tripper   request.RoundTripperServer
	assembler request.SessionAssemblerServer
	addConn   internet.ConnHandler

	streamSettings *internet.MemoryStreamConfig
	addr           net.Address
	port           net.Port
}

func (s server) Listen(ctx context.Context) (net.Listener, error) {
	return transportcommon.ListenWithSecuritySettings(ctx, s.addr, s.port, s.streamSettings)
}

func (s server) AutoImplListener() request.Listener {
	return s
}

func (s server) Close() error {
	if err := s.tripper.Close(); err != nil {
		return newError("failed to close tripper").Base(err)
	}
	return nil
}

func (s server) Addr() net.Addr {
	// Unimplemented
	return nil
}

type serverConnection struct {
	request.Session
}

func (s serverConnection) LocalAddr() gonet.Addr {
	return &net.UnixAddr{Name: "unimplemented"}
}

func (s serverConnection) RemoteAddr() gonet.Addr {
	return &net.UnixAddr{Name: "unimplemented"}
}

func (s serverConnection) SetDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func (s serverConnection) SetReadDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func (s serverConnection) SetWriteDeadline(t time.Time) error {
	// Unimplemented
	return nil
}

func (s server) OnNewSession(ctx context.Context, sess request.Session, opts ...request.SessionOption) error {
	s.addConn(&serverConnection{sess})
	return nil
}

func (s server) SessionReceiver() request.SessionReceiver {
	return s
}

func (s server) TripperReceiver() request.TripperReceiver {
	return s.assembler
}

func listenRequest(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)
	serverAssembly := &server{addConn: addConn}

	assemblerConfigInstance, err := serial.GetInstanceOf(transportConfiguration.Assembler)
	if err != nil {
		return nil, newError("failed to get config instance of assembler").Base(err)
	}
	assembler, err := common.CreateObject(ctx, assemblerConfigInstance)
	if err != nil {
		return nil, newError("failed to create assembler").Base(err)
	}
	if typedAssembler, ok := assembler.(request.SessionAssemblerServer); !ok {
		return nil, newError("failed to type assert assembler to SessionAssemblerServer")
	} else {
		serverAssembly.assembler = typedAssembler
	}

	roundtripperConfigInstance, err := serial.GetInstanceOf(transportConfiguration.Roundtripper)
	if err != nil {
		return nil, newError("failed to get config instance of roundtripper").Base(err)
	}
	roundtripper, err := common.CreateObject(ctx, roundtripperConfigInstance)
	if err != nil {
		return nil, newError("failed to create roundtripper").Base(err)
	}
	if typedRoundtripper, ok := roundtripper.(request.RoundTripperServer); !ok {
		return nil, newError("failed to type assert roundtripper to RoundTripperServer")
	} else {
		serverAssembly.tripper = typedRoundtripper
	}

	serverAssembly.addr = address
	serverAssembly.port = port
	serverAssembly.streamSettings = streamSettings

	serverAssembly.assembler.OnTransportServerAssemblyReady(serverAssembly)
	serverAssembly.tripper.OnTransportServerAssemblyReady(serverAssembly)

	if err := serverAssembly.tripper.Start(); err != nil {
		return nil, newError("failed to start tripper").Base(err)
	}

	return serverAssembly, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, listenRequest))
}
