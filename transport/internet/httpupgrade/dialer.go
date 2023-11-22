package httpupgrade

import (
	"bufio"
	"context"
	"net/http"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"
)

func dialhttpUpgrade(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (net.Conn, error) {
	transportConfiguration := streamSettings.ProtocolSettings.(*Config)

	conn, err := transportcommon.DialWithSecuritySettings(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial request to ", dest).Base(err)
	}
	req, err := http.NewRequest("GET", transportConfiguration.GetNormalizedPath(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Host = transportConfiguration.Host

	err = req.Write(conn)
	if err != nil {
		return nil, err
	}

	// TODO The bufio usage here is unreliable
	resp, err := http.ReadResponse(bufio.NewReader(conn), req) // nolint:bodyclose
	if err != nil {
		return nil, err
	}

	if resp.Status == "101 Switching Protocols" &&
		strings.ToLower(resp.Header.Get("Upgrade")) == "websocket" &&
		strings.ToLower(resp.Header.Get("Connection")) == "upgrade" {
		return conn, nil
	}
	return nil, newError("unrecognized reply")
}

func dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("creating connection to ", dest).WriteToLog(session.ExportIDToError(ctx))

	conn, err := dialhttpUpgrade(ctx, dest, streamSettings)
	if err != nil {
		return nil, newError("failed to dial request to ", dest).Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, dial))
}
