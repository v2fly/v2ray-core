package webrtc

import (
	"net"
	"sync"
	"testing"
	"time"

	pionwebrtc "github.com/pion/webrtc/v4"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

func TestAcceptorListenerSelection(t *testing.T) {
	w := &WebRTC{
		activeListeners: map[string]*activeListenerRuntime{
			"active-a": {config: &LocalWebRTCListener{Tag: "active-a"}},
		},
		systemListeners: map[string]*systemListenerRuntime{
			"sys-a": {config: &LocalWebRTCSystemListener{Tag: "sys-a"}},
			"sys-b": {config: &LocalWebRTCSystemListener{Tag: "sys-b"}},
		},
	}

	t.Run("accept_on_tag", func(t *testing.T) {
		listener, err := w.acceptorListener(&Acceptor{
			Tag:         "acceptor-a",
			AcceptOnTag: "sys-b",
		})
		if err != nil {
			t.Fatal(err)
		}
		if listener != w.systemListeners["sys-b"] {
			t.Fatal("unexpected listener selected")
		}
	})

	t.Run("accept_on_active_tag", func(t *testing.T) {
		listener, err := w.acceptorListener(&Acceptor{
			Tag:         "acceptor-a",
			AcceptOnTag: "active-a",
		})
		if err != nil {
			t.Fatal(err)
		}
		if listener != w.activeListeners["active-a"] {
			t.Fatal("unexpected listener selected")
		}
	})

	t.Run("legacy tag match", func(t *testing.T) {
		listener, err := w.acceptorListener(&Acceptor{
			Tag: "sys-a",
		})
		if err != nil {
			t.Fatal(err)
		}
		if listener != w.systemListeners["sys-a"] {
			t.Fatal("unexpected listener selected")
		}
	})

	t.Run("unknown explicit tag", func(t *testing.T) {
		_, err := w.acceptorListener(&Acceptor{
			Tag:         "acceptor-a",
			AcceptOnTag: "missing",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAcceptorListenerSingleFallback(t *testing.T) {
	only := &systemListenerRuntime{config: &LocalWebRTCSystemListener{Tag: "only"}}
	w := &WebRTC{
		systemListeners: map[string]*systemListenerRuntime{
			"only": only,
		},
	}

	listener, err := w.acceptorListener(&Acceptor{Tag: "acceptor-a"})
	if err != nil {
		t.Fatal(err)
	}
	if listener != only {
		t.Fatal("unexpected listener selected")
	}
}

func TestAcceptorSessionPortBlossomRepeatsUntilConnected(t *testing.T) {
	oldInterval := portBlossomRepeatInterval
	portBlossomRepeatInterval = 10 * time.Millisecond
	defer func() {
		portBlossomRepeatInterval = oldInterval
	}()

	listener := &countingListenerRuntime{
		counts: make(map[string]int),
		limit:  time.Second,
	}
	session := &acceptorSession{
		owner: &acceptorRuntime{
			tag:      "acceptor-a",
			listener: listener,
		},
		sessionID:          []byte("session-a"),
		portBlossomIPs:     make(map[string]v2net.IP),
		portBlossomRunning: true,
		portBlossomStop:    make(chan struct{}),
	}
	session.portBlossomIPs["198.51.100.10"] = append(v2net.IP(nil), net.ParseIP("198.51.100.10")...)
	session.portBlossomIPs["203.0.113.20"] = append(v2net.IP(nil), net.ParseIP("203.0.113.20")...)

	go session.portBlossomLoop()

	waitForPortBlastCount(t, listener, "198.51.100.10", 2)
	waitForPortBlastCount(t, listener, "203.0.113.20", 2)

	beforeStop := listener.snapshotCounts()
	session.setPeerConnected()
	time.Sleep(4 * portBlossomRepeatInterval)
	afterStop := listener.snapshotCounts()

	for ip, before := range beforeStop {
		if after := afterStop[ip]; after != before {
			t.Fatalf("port blossom for %s continued after connected: before=%d after=%d", ip, before, after)
		}
	}
}

func TestLocalCandidateSendGateDisabledAllowsImmediateSend(t *testing.T) {
	gate := newLocalCandidateSendGate(false, remoteCandidateGatheringWorkaroundDelay, func() time.Time {
		return time.Unix(100, 0)
	})

	if !gate.AllowSend() {
		t.Fatal("expected immediate send when workaround is disabled")
	}
	if gate.StartCountdown() {
		t.Fatal("disabled gate should not start a countdown")
	}
	if got := gate.NextPollDelay(signalPollInterval); got != signalPollInterval {
		t.Fatalf("unexpected poll delay: got=%v want=%v", got, signalPollInterval)
	}
}

func TestLocalCandidateSendGateHoldbackLifecycle(t *testing.T) {
	current := time.Unix(100, 0)
	gate := newLocalCandidateSendGate(true, 500*time.Millisecond, func() time.Time {
		return current
	})

	if gate.AllowSend() {
		t.Fatal("expected send to be blocked before remote gathering completes")
	}
	if got := gate.NextPollDelay(signalPollInterval); got != signalPollInterval {
		t.Fatalf("unexpected delay before countdown start: got=%v want=%v", got, signalPollInterval)
	}
	if !gate.StartCountdown() {
		t.Fatal("expected countdown to start")
	}
	if gate.StartCountdown() {
		t.Fatal("countdown should only start once")
	}
	if gate.AllowSend() {
		t.Fatal("expected send to remain blocked during holdback")
	}
	if got := gate.NextPollDelay(signalPollInterval); got != 500*time.Millisecond {
		t.Fatalf("unexpected delay during holdback: got=%v want=%v", got, 500*time.Millisecond)
	}

	current = current.Add(300 * time.Millisecond)
	if gate.AllowSend() {
		t.Fatal("expected send to remain blocked until countdown expires")
	}
	if got := gate.NextPollDelay(signalPollInterval); got != 200*time.Millisecond {
		t.Fatalf("unexpected delay before release: got=%v want=%v", got, 200*time.Millisecond)
	}

	current = current.Add(200 * time.Millisecond)
	if !gate.AllowSend() {
		t.Fatal("expected send to be allowed after countdown expires")
	}
	if got := gate.NextPollDelay(signalPollInterval); got != signalPollInterval {
		t.Fatalf("unexpected delay after release: got=%v want=%v", got, signalPollInterval)
	}
}

type countingListenerRuntime struct {
	mu     sync.Mutex
	counts map[string]int
	limit  time.Duration
}

func (l *countingListenerRuntime) Tag() string {
	return "counting"
}

func (l *countingListenerRuntime) NewPeerAPI() (*pionwebrtc.API, pionwebrtc.Configuration, error) {
	return nil, pionwebrtc.Configuration{}, nil
}

func (l *countingListenerRuntime) RequestPortBlossom() bool {
	return false
}

func (l *countingListenerRuntime) AcceptPortBlossom() bool {
	return true
}

func (l *countingListenerRuntime) PortBlossomDuration() time.Duration {
	if l.limit <= 0 {
		return defaultPortBlossomDuration
	}
	return l.limit
}

func (l *countingListenerRuntime) RequireRemoteFinishCandidateGatheringBeforeSendingOwnCandidateWorkaround() bool {
	return false
}

func (l *countingListenerRuntime) BlastPorts(ip net.IP) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[ip.String()]++
	return nil
}

func (l *countingListenerRuntime) Close() error {
	return nil
}

func (l *countingListenerRuntime) snapshotCounts() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()

	counts := make(map[string]int, len(l.counts))
	for ip, count := range l.counts {
		counts[ip] = count
	}
	return counts
}

func waitForPortBlastCount(t *testing.T, listener *countingListenerRuntime, ip string, want int) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if listener.snapshotCounts()[ip] >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for port blossom on %s to reach %d, got %d", ip, want, listener.snapshotCounts()[ip])
}

func TestAcceptorSessionPortBlossomStopsAfterTimeout(t *testing.T) {
	oldInterval := portBlossomRepeatInterval
	portBlossomRepeatInterval = 10 * time.Millisecond
	defer func() {
		portBlossomRepeatInterval = oldInterval
	}()

	listener := &countingListenerRuntime{
		counts: make(map[string]int),
		limit:  35 * time.Millisecond,
	}
	session := &acceptorSession{
		owner: &acceptorRuntime{
			tag:      "acceptor-a",
			listener: listener,
		},
		sessionID:          []byte("session-b"),
		portBlossomIPs:     make(map[string]v2net.IP),
		portBlossomRunning: true,
		portBlossomStop:    make(chan struct{}),
	}
	session.portBlossomIPs["198.51.100.10"] = append(v2net.IP(nil), net.ParseIP("198.51.100.10")...)

	go session.portBlossomLoop()

	waitForPortBlastCount(t, listener, "198.51.100.10", 2)
	time.Sleep(6 * portBlossomRepeatInterval)
	beforeStop := listener.snapshotCounts()
	time.Sleep(4 * portBlossomRepeatInterval)
	afterStop := listener.snapshotCounts()

	if got := afterStop["198.51.100.10"]; got != beforeStop["198.51.100.10"] {
		t.Fatalf("port blossom continued after timeout: before=%d after=%d", beforeStop["198.51.100.10"], got)
	}
}
