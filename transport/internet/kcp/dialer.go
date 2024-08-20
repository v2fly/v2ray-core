package kcp

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/dice"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

var globalConv = uint32(dice.RollUint16())

func fetchInput(_ context.Context, input io.Reader, reader PacketReader, conn *Connection) {
	cache := make(chan *buf.Buffer, 1024)
	go func() {
		for {
			payload := buf.New()
			if _, err := payload.ReadFrom(input); err != nil {
				payload.Release()
				close(cache)
				return
			}
			select {
			case cache <- payload:
			default:
				payload.Release()
			}
		}
	}()

	for payload := range cache {
		segments := reader.Read(payload.Bytes())
		payload.Release()
		if len(segments) > 0 {
			conn.Input(segments)
		}
	}
}

// DialKCP dials a new KCP connections to the specific destination.
func DialKCP(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	dest.Network = net.Network_UDP
	newError("dialing mKCP to ", dest).WriteToLog()

	transportEnvironment := envctx.EnvironmentFromContext(ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.Dialer()

	rawConn, err := dialer.Dial(ctx, nil, dest, streamSettings.SocketSettings)
	if err != nil {
		return nil, newError("failed to dial to dest: ", err).AtWarning().Base(err)
	}

	kcpSettings := streamSettings.ProtocolSettings.(*Config)

	header, err := kcpSettings.GetPackerHeader()
	if err != nil {
		return nil, newError("failed to create packet header").Base(err)
	}
	security, err := kcpSettings.GetSecurity()
	if err != nil {
		return nil, newError("failed to create security").Base(err)
	}
	reader := &KCPPacketReader{
		Header:   header,
		Security: security,
	}
	writer := &KCPPacketWriter{
		Header:   header,
		Security: security,
		Writer:   rawConn,
	}

	conv := uint16(atomic.AddUint32(&globalConv, 1))
	session := NewConnection(ConnMetadata{
		LocalAddr:    rawConn.LocalAddr(),
		RemoteAddr:   rawConn.RemoteAddr(),
		Conversation: conv,
	}, writer, rawConn, kcpSettings)

	go fetchInput(ctx, rawConn, reader, session)

	var iConn internet.Connection = session

	if config := tls.ConfigFromStreamSettings(streamSettings); config != nil {
		iConn = tls.Client(iConn, config.GetTLSConfig(tls.WithDestination(dest)))
	}

	return iConn, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, DialKCP))
}
