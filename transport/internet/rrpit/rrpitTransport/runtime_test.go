package rrpitTransport

import (
	"bytes"
	"context"
	"io"
	gonet "net"
	"sync"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSessionManager"
)

type testClientIdentityConn struct {
	gonet.Conn
	identity []byte
}

func (c *testClientIdentityConn) ClientIdentity() []byte {
	return append([]byte(nil), c.identity...)
}

func TestTransportSessionIdentityRoundTrip(t *testing.T) {
	want := transportSessionID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	conn := &testClientIdentityConn{identity: encodeTransportSessionIdentity(want)}

	got, err := readTransportSessionID(conn)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("unexpected session id: got %x want %x", got, want)
	}
}

func TestTransportSessionIdentityRejectsInvalidIdentity(t *testing.T) {
	tests := []struct {
		name     string
		identity []byte
	}{
		{name: "empty"},
		{name: "short", identity: []byte("rrpit")},
		{name: "bad magic", identity: append([]byte("WRNG"), append([]byte{transportIdentityVersion}, bytes.Repeat([]byte{1}, len(transportSessionID{}))...)...)},
		{name: "bad version", identity: append(append([]byte(nil), transportIdentityMagic[:]...), append([]byte{transportIdentityVersion + 1}, bytes.Repeat([]byte{2}, len(transportSessionID{}))...)...)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn := &testClientIdentityConn{identity: test.identity}
			if _, err := readTransportSessionID(conn); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestBuildBidirectionalSessionConfigIncludesReconstructionSettings(t *testing.T) {
	config := buildBidirectionalSessionConfig(&Config{
		Lane: &LaneSetting{
			ShardSize:            1400,
			MaxDataShardsPerLane: 6,
			MaxBufferedLanes:     8,
		},
		Session: &SessionSetting{
			OddChannelIds:                  true,
			MaxRewindableTimestampNum:      9,
			MaxRewindableControlMessageNum: 10,
			Reconstruction: &SessionReconstructionSetting{
				InitialRepairShardRatio:                       1.25,
				LaneRepairWeight:                              []float32{0.5, 0.25},
				SecondaryRepairShardRatio:                     0.75,
				TimeResendSecondaryRepairShard:                3,
				StaleLaneFinalizedAgeThresholdTicks:           7,
				StaleLaneProgressStallThresholdTicks:          8,
				SecondaryRepairMinBurst:                       2,
				AlwaysRestrictSourceDataWhenOldestLaneStalled: true,
			},
		},
		SessionMgr: &SessionManagerSetting{
			TimestampInterval: 11,
		},
	})

	want := rriptMonoDirectionSession.SessionTxReconstructionConfig{
		InitialRepairShardRatio:                       1.25,
		LaneRepairWeight:                              []float64{0.5, 0.25},
		SecondaryRepairShardRatio:                     0.75,
		TimeResendSecondaryRepairShard:                3,
		StaleLaneFinalizedAgeThresholdTicks:           7,
		StaleLaneProgressStallThresholdTicks:          8,
		SecondaryRepairMinBurst:                       2,
		AlwaysRestrictSourceDataWhenOldestLaneStalled: true,
	}
	got := config.Tx.Reconstruction
	if got.InitialRepairShardRatio != want.InitialRepairShardRatio ||
		got.SecondaryRepairShardRatio != want.SecondaryRepairShardRatio ||
		got.TimeResendSecondaryRepairShard != want.TimeResendSecondaryRepairShard ||
		got.StaleLaneFinalizedAgeThresholdTicks != want.StaleLaneFinalizedAgeThresholdTicks ||
		got.StaleLaneProgressStallThresholdTicks != want.StaleLaneProgressStallThresholdTicks ||
		got.SecondaryRepairMinBurst != want.SecondaryRepairMinBurst ||
		len(got.LaneRepairWeight) != len(want.LaneRepairWeight) {
		t.Fatalf("unexpected reconstruction config: %+v", got)
	}
	for i, weight := range want.LaneRepairWeight {
		if got.LaneRepairWeight[i] != weight {
			t.Fatalf("unexpected lane repair weight[%d]: got %v want %v", i, got.LaneRepairWeight[i], weight)
		}
	}
}

func TestPersistentClientSessionKeepsTransportAliveUntilIdle(t *testing.T) {
	clientConn, serverConn := gonet.Pipe()
	defer serverConn.Close()

	var (
		closeCalls int
		closeMu    sync.Mutex
	)

	session := &persistentClientSession{
		owner: &transportSession{
			localAddr:  &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10001},
			remoteAddr: &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10002},
		},
		openStream: func(rrpitBidirectionalSessionManager.SessionName) (gonet.Conn, error) {
			return clientConn, nil
		},
		closeSession: func() error {
			closeMu.Lock()
			closeCalls++
			closeMu.Unlock()
			return nil
		},
		idleTimeout: 20 * time.Millisecond,
	}

	conn, err := session.OpenConnection(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if conn.LocalAddr().String() != "127.0.0.1:10001" {
		t.Fatalf("unexpected local addr: %v", conn.LocalAddr())
	}
	if conn.RemoteAddr().String() != "127.0.0.1:10002" {
		t.Fatalf("unexpected remote addr: %v", conn.RemoteAddr())
	}

	if err := conn.Close(); err != nil && err != io.ErrClosedPipe {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Millisecond)
	closeMu.Lock()
	if closeCalls != 0 {
		closeMu.Unlock()
		t.Fatalf("transport closed too early: %d", closeCalls)
	}
	closeMu.Unlock()

	time.Sleep(40 * time.Millisecond)
	closeMu.Lock()
	defer closeMu.Unlock()
	if closeCalls != 1 {
		t.Fatalf("unexpected transport close count: %d", closeCalls)
	}
}

func TestPersistentClientSessionOpenConnectionTracksContext(t *testing.T) {
	clientConn, serverConn := gonet.Pipe()
	defer serverConn.Close()

	var (
		closeCalls int
		closeMu    sync.Mutex
	)

	session := &persistentClientSession{
		owner: &transportSession{
			localAddr:  &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10011},
			remoteAddr: &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10012},
		},
		openStream: func(rrpitBidirectionalSessionManager.SessionName) (gonet.Conn, error) {
			return clientConn, nil
		},
		closeSession: func() error {
			closeMu.Lock()
			closeCalls++
			closeMu.Unlock()
			return nil
		},
		idleTimeout: 20 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn, err := session.OpenConnection(ctx)
	if err != nil {
		t.Fatal(err)
	}

	getter, ok := conn.(interface{ GetConnectionContext() context.Context })
	if !ok {
		t.Fatal("connection does not expose connection context")
	}
	if getter.GetConnectionContext() != ctx {
		t.Fatal("unexpected connection context")
	}

	cancel()

	time.Sleep(5 * time.Millisecond)
	closeMu.Lock()
	if closeCalls != 0 {
		closeMu.Unlock()
		t.Fatalf("transport closed too early: %d", closeCalls)
	}
	closeMu.Unlock()

	time.Sleep(40 * time.Millisecond)
	closeMu.Lock()
	defer closeMu.Unlock()
	if closeCalls != 1 {
		t.Fatalf("unexpected transport close count after context cancel: %d", closeCalls)
	}
}
