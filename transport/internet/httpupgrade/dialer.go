package httpupgrade

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
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

	dialer := func(earlyData []byte) (net.Conn, io.Reader, error) {
		conn, err := transportcommon.DialWithSecuritySettings(ctx, dest, streamSettings)
		if err != nil {
			return nil, nil, newError("failed to dial request to ", dest).Base(err)
		}
		req, err := http.NewRequest("GET", transportConfiguration.GetNormalizedPath(), nil)
		if err != nil {
			return nil, nil, err
		}

		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Host = transportConfiguration.Host

		if transportConfiguration.Header != nil {
			for _, value := range transportConfiguration.Header {
				req.Header.Set(value.Key, value.Value)
			}
		}

		earlyDataSize := len(earlyData)
		if earlyDataSize > int(transportConfiguration.MaxEarlyData) {
			earlyDataSize = int(transportConfiguration.MaxEarlyData)
		}

		if len(earlyData) > 0 {
			if transportConfiguration.EarlyDataHeaderName == "" {
				return nil, nil, newError("EarlyDataHeaderName is not set")
			}
			req.Header.Set(transportConfiguration.EarlyDataHeaderName, base64.URLEncoding.EncodeToString(earlyData))
		}

		err = req.Write(conn)
		if err != nil {
			return nil, nil, err
		}

		if earlyData != nil && len(earlyData[earlyDataSize:]) > 0 {
			_, err = conn.Write(earlyData[earlyDataSize:])
			if err != nil {
				return nil, nil, newError("failed to finish write early data").Base(err)
			}
		}

		bufferedConn := bufio.NewReader(conn)
		resp, err := http.ReadResponse(bufferedConn, req) // nolint:bodyclose
		if err != nil {
			return nil, nil, err
		}

		if resp.Status == "101 Switching Protocols" &&
			strings.ToLower(resp.Header.Get("Upgrade")) == "websocket" &&
			strings.ToLower(resp.Header.Get("Connection")) == "upgrade" {
			earlyReplyReader := io.LimitReader(bufferedConn, int64(bufferedConn.Buffered()))
			return conn, earlyReplyReader, nil
		}

		return nil, nil, newError("unrecognized reply")
	}

	if transportConfiguration.MaxEarlyData == 0 {
		conn, earlyReplyReader, err := dialer(nil)
		if err != nil {
			return nil, err
		}
		remoteAddr := conn.RemoteAddr()

		return newConnectionWithPendingRead(conn, remoteAddr, earlyReplyReader), nil
	}

	return newConnectionWithDelayedDial(dialer), nil
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
