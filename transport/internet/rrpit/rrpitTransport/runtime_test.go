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
			ShardSize:                  1400,
			MaxDataShardsPerLane:       6,
			MaxBufferedLanes:           8,
			RemoteMaxDataShardsPerLane: 5,
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
	if config.Rx.RemoteMaxDataShardsPerLane != 5 {
		t.Fatalf("unexpected remote max data shards per lane: got %d want 5", config.Rx.RemoteMaxDataShardsPerLane)
	}
}

func TestBuildSmuxConfigClampsDefaultFrameSizeToMessageBudget(t *testing.T) {
	config, err := buildSmuxConfig(&AdaptorSetting{}, 1198)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := config.MaxFrameSize, 1168; got != want {
		t.Fatalf("unexpected default smux max frame size: got %d want %d", got, want)
	}

	explicit, err := buildSmuxConfig(&AdaptorSetting{MaxFrameSize: 2048}, 1198)
	if err != nil {
		t.Fatal(err)
	}
	if explicit.MaxFrameSize != 2048 {
		t.Fatalf("unexpected explicit smux max frame size: got %d want 2048", explicit.MaxFrameSize)
	}
}

func TestBuildConnectionPersistencePolicy(t *testing.T) {
	defaults := buildConnectionPersistencePolicy(&Config{})
	if defaults.DisconnectedSessionRetention != 0 {
		t.Fatalf("unexpected default disconnected retention: %v", defaults.DisconnectedSessionRetention)
	}
	if defaults.ReconnectRetryInterval != 0 {
		t.Fatalf("unexpected default reconnect retry interval: %v", defaults.ReconnectRetryInterval)
	}
	if defaults.IdleTimeout != rrpitClientSessionIdleTimeout {
		t.Fatalf("unexpected default idle timeout: %v", defaults.IdleTimeout)
	}
	if defaults.RemoteControlInactivityTimeout != 0 {
		t.Fatalf("unexpected default remote control inactivity timeout: %v", defaults.RemoteControlInactivityTimeout)
	}
	if defaults.KeepTransportSessionWithoutStreams {
		t.Fatal("did not expect keep-without-streams default to be enabled")
	}

	policy := buildConnectionPersistencePolicy(&Config{
		Persistence: &ConnectionPersistenceSetting{
			DisconnectedSessionRetention:       int64(3 * time.Second),
			ReconnectRetryInterval:             int64(250 * time.Millisecond),
			KeepTransportSessionWithoutStreams: true,
			IdleTimeout:                        int64(7 * time.Second),
			RemoteControlInactivityTimeout:     int64(9 * time.Second),
		},
	})
	if policy.DisconnectedSessionRetention != 3*time.Second {
		t.Fatalf("unexpected disconnected retention: %v", policy.DisconnectedSessionRetention)
	}
	if policy.ReconnectRetryInterval != 250*time.Millisecond {
		t.Fatalf("unexpected reconnect retry interval: %v", policy.ReconnectRetryInterval)
	}
	if policy.IdleTimeout != 7*time.Second {
		t.Fatalf("unexpected idle timeout: %v", policy.IdleTimeout)
	}
	if policy.RemoteControlInactivityTimeout != 9*time.Second {
		t.Fatalf("unexpected remote control inactivity timeout: %v", policy.RemoteControlInactivityTimeout)
	}
	if !policy.KeepTransportSessionWithoutStreams {
		t.Fatal("expected keep-without-streams to be enabled")
	}
}

func TestChannelReadIdleTimeout(t *testing.T) {
	if got := channelReadIdleTimeout(resolvedChannel{}); got != 0 {
		t.Fatalf("unexpected zero-config timeout: %v", got)
	}
	if got := channelReadIdleTimeout(resolvedChannel{dtls: &DTLSUDPChannel{NoIncomingMessageTimeout: int64(250 * time.Millisecond)}}); got != 250*time.Millisecond {
		t.Fatalf("unexpected configured timeout: %v", got)
	}
	if got := channelReadIdleTimeout(resolvedChannel{dtls: &DTLSUDPChannel{NoIncomingMessageTimeout: -1}}); got != 0 {
		t.Fatalf("unexpected negative timeout handling: %v", got)
	}
}

func TestChannelReadFailureMessageUsesTimeoutWording(t *testing.T) {
	timeoutErr := &timeoutOnlyError{}
	if got := channelReadFailureMessage("failed to read channel packet", timeoutErr); got != "channel considered dead after no incoming message within configured timeout" {
		t.Fatalf("unexpected timeout failure message: %q", got)
	}
	if got := channelReadFailureMessage("failed to read channel packet", io.EOF); got != "failed to read channel packet" {
		t.Fatalf("unexpected non-timeout failure message: %q", got)
	}
}

func TestReadChannelPacketPreservesPacketBoundaries(t *testing.T) {
	conn := &testPacketConn{
		packets: [][]byte{
			[]byte("first"),
			[]byte("second"),
		},
	}

	buffer := make([]byte, 16)
	first, err := readChannelPacket(conn, 50*time.Millisecond, buffer)
	if err != nil {
		t.Fatal(err)
	}
	second, err := readChannelPacket(conn, 50*time.Millisecond, buffer)
	if err != nil {
		t.Fatal(err)
	}

	if string(first) != "first" {
		t.Fatalf("unexpected first packet: %q", string(first))
	}
	if string(second) != "second" {
		t.Fatalf("unexpected second packet: %q", string(second))
	}
	if conn.readDeadline.IsZero() {
		t.Fatal("expected read deadline to be set")
	}
	buffer[0] = 'x'
	if string(first) != "first" {
		t.Fatalf("first packet was not copied: %q", string(first))
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

func TestPersistentClientSessionCanStayOpenWithoutStreams(t *testing.T) {
	clientConn, serverConn := gonet.Pipe()
	defer serverConn.Close()

	var (
		closeCalls int
		closeMu    sync.Mutex
	)

	session := &persistentClientSession{
		owner: &transportSession{
			localAddr:  &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10021},
			remoteAddr: &gonet.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 10022},
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
		idleTimeout:                        20 * time.Millisecond,
		keepTransportSessionWithoutStreams: true,
	}

	conn, err := session.OpenConnection(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil && err != io.ErrClosedPipe {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	closeMu.Lock()
	defer closeMu.Unlock()
	if closeCalls != 0 {
		t.Fatalf("expected transport to stay open without streams, got close count %d", closeCalls)
	}
}

func TestTransportSessionDisconnectedRetentionDelaysClose(t *testing.T) {
	var (
		closeCalls int
		closeMu    sync.Mutex
	)

	session := &transportSession{
		persistence: connectionPersistencePolicy{
			DisconnectedSessionRetention: 20 * time.Millisecond,
		},
		onClose: func() {
			closeMu.Lock()
			closeCalls++
			closeMu.Unlock()
		},
	}

	session.mu.Lock()
	closeImmediately := session.scheduleDisconnectedTimerLocked()
	session.mu.Unlock()
	if closeImmediately {
		t.Fatal("expected disconnect retention to delay close")
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
		t.Fatalf("unexpected transport close count after retention expiry: %d", closeCalls)
	}
}

func TestTransportSessionCloseDoesNotDeadlockWithAutoTickAndNoChannels(t *testing.T) {
	session, err := newTransportSession(
		"server",
		transportSessionID{},
		&Config{
			SessionMgr: &SessionManagerSetting{
				TimestampInterval: int64(time.Millisecond),
			},
			Persistence: &ConnectionPersistenceSetting{
				DisconnectedSessionRetention: int64(time.Second),
			},
		},
		false,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- session.Close()
	}()

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("transport session close timed out")
	}
}

type timeoutOnlyError struct{}

func (*timeoutOnlyError) Error() string { return "timeout" }
func (*timeoutOnlyError) Timeout() bool { return true }

type testPacketConn struct {
	gonet.Conn
	packets      [][]byte
	readErr      error
	readDeadline time.Time
}

func (c *testPacketConn) Read(b []byte) (int, error) {
	if c.readErr != nil {
		return 0, c.readErr
	}
	if len(c.packets) == 0 {
		return 0, io.EOF
	}
	packet := c.packets[0]
	c.packets = c.packets[1:]
	if len(b) < len(packet) {
		return 0, io.ErrShortBuffer
	}
	copy(b, packet)
	return len(packet), nil
}

func (c *testPacketConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *testPacketConn) Close() error {
	return nil
}

func (c *testPacketConn) LocalAddr() gonet.Addr {
	return nil
}

func (c *testPacketConn) RemoteAddr() gonet.Addr {
	return nil
}

func (c *testPacketConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *testPacketConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *testPacketConn) SetWriteDeadline(time.Time) error {
	return nil
}
