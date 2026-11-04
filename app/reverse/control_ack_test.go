package reverse

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/mux"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

func TestBridgeAcknowledgesOnlyNewControl(t *testing.T) {
	incoming, incomingWriter := pipe.New(pipe.WithoutSizeLimit())
	outgoing, outgoingWriter := pipe.New(pipe.WithoutSizeLimit())
	bridge := &BridgeWorker{}
	bridge.handleInternalConn(transport.Link{Reader: incoming, Writer: outgoingWriter})
	t.Cleanup(func() {
		incoming.Interrupt()
		outgoing.Interrupt()
	})

	legacy, err := proto.Marshal(&Control{State: Control_ACTIVE, Random: []byte("legacy")})
	if err != nil {
		t.Fatal(err)
	}
	if err := incomingWriter.WriteMultiBuffer(buf.MergeBytes(nil, legacy)); err != nil {
		t.Fatal(err)
	}
	mb, err := outgoing.ReadMultiBufferTimeout(100 * time.Millisecond)
	buf.ReleaseMulti(mb)
	if err != buf.ErrReadTimeout {
		t.Fatalf("legacy portal must not receive unsolicited replies, got %v", err)
	}

	probe, err := proto.Marshal(&Control{State: Control_ACTIVE, HeartbeatId: 7})
	if err != nil {
		t.Fatal(err)
	}
	if err := incomingWriter.WriteMultiBuffer(buf.MergeBytes(nil, probe)); err != nil {
		t.Fatal(err)
	}
	mb, err = outgoing.ReadMultiBufferTimeout(3 * time.Second)
	if err != nil || len(mb) != 1 {
		buf.ReleaseMulti(mb)
		t.Fatalf("no control acknowledgement: %v", err)
	}
	var reply Control
	err = proto.Unmarshal(mb[0].Bytes(), &reply)
	buf.ReleaseMulti(mb)
	if err != nil || reply.HeartbeatId != 7 || reply.State != Control_ACTIVE {
		t.Fatalf("unexpected reply: %v, %v", &reply, err)
	}
	bridge.controlAccess.Lock()
	observed := !bridge.lastControl.IsZero()
	bridge.controlAccess.Unlock()
	if !observed {
		t.Fatal("bridge did not record the opted-in portal heartbeat")
	}
	incomingWriter.Close()
	deadline := time.Now().Add(3 * time.Second)
	for bridgeState(bridge) != Control_DRAIN {
		if time.Now().After(deadline) {
			t.Fatal("bridge stayed ACTIVE after its control stream ended")
		}
		time.Sleep(time.Millisecond)
	}
}

func bridgeState(w *BridgeWorker) Control_State {
	w.controlAccess.Lock()
	defer w.controlAccess.Unlock()
	return w.state
}

func newAckTestClient(t *testing.T) *mux.ClientWorker {
	t.Helper()
	input, _ := pipe.New(pipe.WithoutSizeLimit())
	client, err := mux.NewClientWorker(transport.Link{Reader: input, Writer: buf.Discard}, mux.ClientStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		client.Close()
		input.Interrupt()
	})
	return client
}

func TestPortalOnlyExpiresAcknowledgedWorker(t *testing.T) {
	oldClient := newAckTestClient(t)
	controlReader, controlWriter := pipe.New(pipe.WithoutSizeLimit())
	ackReader, _ := pipe.New(pipe.WithoutSizeLimit())
	t.Cleanup(func() {
		controlReader.Interrupt()
		ackReader.Interrupt()
	})
	old := &PortalWorker{
		client:  oldClient,
		writer:  controlWriter,
		reader:  ackReader,
		control: &task.Periodic{},
	}
	picker := &StaticMuxPicker{}
	picker.AddWorker(old)

	// A legacy bridge ignores the new field. It must remain usable.
	old.heartbeat.probeID.Store(1)
	if old.expireIfUnhealthy(time.Now().Add(time.Hour)) || oldClient.Closed() {
		t.Fatal("worker without an acknowledgement was rejected")
	}

	freshClient := newAckTestClient(t)
	picker.AddWorker(&PortalWorker{client: freshClient})
	start := time.Now()
	old.recordAck(2, start) // No heartbeat 2 was sent; ignore it.
	if !old.heartbeat.lastAckTime.IsZero() {
		t.Fatal("accepted an acknowledgement for a future heartbeat")
	}
	old.recordAck(1, start)
	old.recordAck(1, start.Add(time.Second)) // Replays must not refresh the timer.
	if old.heartbeat.lastAckTime != start {
		t.Fatal("replayed acknowledgement refreshed the timer")
	}
	if old.expireIfUnhealthy(start.Add(controlSilenceTimeout - time.Millisecond)) {
		t.Fatal("worker was expired before the deadline")
	}
	old.heartbeat.probeID.Store(2)
	freshAck := start.Add(controlSilenceTimeout / 2)
	old.recordAck(2, freshAck)
	if old.expireIfUnhealthy(start.Add(controlSilenceTimeout + time.Millisecond)) {
		t.Fatal("a fresh acknowledgement did not extend the deadline")
	}
	if !old.expireIfUnhealthy(freshAck.Add(controlSilenceTimeout+time.Millisecond)) || !oldClient.Closed() {
		t.Fatal("acknowledged worker did not expire after missing acknowledgements")
	}
	picked, err := picker.PickAvailable()
	if err != nil || picked != freshClient {
		t.Fatalf("picker reused the expired worker: picked=%p err=%v", picked, err)
	}
}

func TestPickerDistinguishesGracefulDrainAndFailure(t *testing.T) {
	old := newAckTestClient(t)
	picker := &StaticMuxPicker{}
	draining := &PortalWorker{client: old}
	draining.state.Store(portalWorkerDraining)
	picker.AddWorker(draining)
	// Graceful rollover must keep serving until the replacement arrives.
	if picked, err := picker.PickAvailable(); err != nil || picked != old {
		t.Fatalf("lost the only graceful worker: picked=%p err=%v", picked, err)
	}

	failed := newAckTestClient(t)
	unhealthy := &PortalWorker{client: failed}
	unhealthy.state.Store(portalWorkerFailed)
	failedPicker := &StaticMuxPicker{}
	failedPicker.AddWorker(unhealthy)
	if picked, err := failedPicker.PickAvailable(); err == nil || picked != nil {
		t.Fatalf("selected the failed worker: picked=%p err=%v", picked, err)
	}
	picker.AddWorker(unhealthy)
	if picked, err := picker.PickAvailable(); err != nil || picked != old {
		t.Fatalf("selected a failed worker over a graceful one: picked=%p err=%v", picked, err)
	}

	fresh := newAckTestClient(t)
	picker.AddWorker(&PortalWorker{client: fresh})
	if picked, err := picker.PickAvailable(); err != nil || picked != fresh {
		t.Fatalf("did not select the fresh worker: picked=%p err=%v", picked, err)
	}
}

func TestPortalDrainingDoesNotExpire(t *testing.T) {
	client := newAckTestClient(t)
	controlReader, controlWriter := pipe.New(pipe.WithoutSizeLimit())
	ackReader, _ := pipe.New(pipe.WithoutSizeLimit())
	t.Cleanup(func() {
		controlReader.Interrupt()
		ackReader.Interrupt()
	})
	w := &PortalWorker{client: client, writer: controlWriter, reader: ackReader, control: &task.Periodic{}}
	w.heartbeat.probeID.Store(1)
	w.recordAck(1, time.Now())
	w.state.Store(portalWorkerDraining)
	if w.expireIfUnhealthy(time.Now().Add(controlSilenceTimeout+time.Minute)) || client.Closed() {
		t.Fatal("a worker draining existing sessions was aborted")
	}
}

type controlledDropWriter struct {
	writer  buf.Writer
	discard atomic.Bool
	once    sync.Once
	dropped chan struct{}
}

func (w *controlledDropWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if !w.discard.Load() {
		return w.writer.WriteMultiBuffer(mb)
	}
	buf.ReleaseMulti(mb)
	w.once.Do(func() { close(w.dropped) })
	return nil
}

type controlTestPair struct {
	ctx      context.Context
	portal   *PortalWorker
	bridge   *BridgeWorker
	client   *mux.ClientWorker
	server   *mux.ServerWorker
	requests *controlledDropWriter
	replies  *controlledDropWriter
}

func newControlTestPair(t *testing.T) *controlTestPair {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	toBridgeReader, toBridgeWriter := pipe.New(pipe.WithoutSizeLimit())
	toPortalReader, toPortalWriter := pipe.New(pipe.WithoutSizeLimit())
	requests := &controlledDropWriter{writer: toBridgeWriter, dropped: make(chan struct{})}
	replies := &controlledDropWriter{writer: toPortalWriter, dropped: make(chan struct{})}
	client, err := mux.NewClientWorker(transport.Link{Reader: toPortalReader, Writer: requests}, mux.ClientStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	bridge := &BridgeWorker{}
	server, err := mux.NewServerWorker(ctx, bridge, &transport.Link{Reader: toBridgeReader, Writer: replies})
	if err != nil {
		t.Fatal(err)
	}
	bridge.worker = server
	portal, err := NewPortalWorker(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		portal.control.Close()
		client.Close()
		server.Close()
		cancel()
		toBridgeReader.Interrupt()
		toPortalReader.Interrupt()
		toPortalWriter.Interrupt()
		toBridgeWriter.Interrupt()
	})
	return &controlTestPair{ctx: ctx, portal: portal, bridge: bridge, client: client, server: server, requests: requests, replies: replies}
}

func waitForPortalAck(t *testing.T, portal *PortalWorker) {
	t.Helper()
	deadline := time.Now().Add(3 * controlHeartbeatInterval)
	for {
		portal.heartbeat.mu.Lock()
		confirmed := !portal.heartbeat.lastAckTime.IsZero()
		portal.heartbeat.mu.Unlock()
		if confirmed {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("bridge did not acknowledge the portal's control heartbeat")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestPortalExpiresAfterLostReplies(t *testing.T) {
	pair := newControlTestPair(t)
	waitForPortalAck(t, pair.portal)

	// Keep both transport pipes open, but drop Bridge-to-Portal traffic.
	// Only the independent watchdog can now make this worker unavailable.
	pair.replies.discard.Store(true)
	select {
	case <-pair.replies.dropped:
	case <-time.After(3 * controlHeartbeatInterval):
		t.Fatal("no acknowledgement was dropped")
	}
	if pair.client.Closed() {
		t.Fatal("worker closed before the timeout despite an open transport")
	}
	pair.portal.heartbeat.mu.Lock()
	pair.portal.heartbeat.lastAckID = pair.portal.heartbeat.probeID.Load() // Exclude earlier buffered replies.
	pair.portal.heartbeat.lastAckTime = time.Now().Add(-controlSilenceTimeout - time.Second)
	pair.portal.heartbeat.mu.Unlock()
	deadline := time.Now().Add(3 * controlHeartbeatInterval)
	for !pair.client.Closed() {
		if time.Now().After(deadline) {
			t.Fatal("watchdog did not close the unresponsive worker")
		}
		time.Sleep(time.Millisecond)
	}
	if state := pair.portal.state.Load(); state != portalWorkerFailed {
		t.Fatalf("worker closed without a health timeout: state=%d", state)
	}
	picker := &StaticMuxPicker{}
	picker.AddWorker(pair.portal)
	if picked, err := picker.PickAvailable(); err == nil || picked != nil {
		t.Fatalf("picker selected the expired worker: picked=%p err=%v", picked, err)
	}
}

func TestBridgeMonitorReplacesWorkerAfterOneWayStall(t *testing.T) {
	pair := newControlTestPair(t)
	waitForPortalAck(t, pair.portal)
	manager, err := NewBridge(pair.ctx, &BridgeConfig{Domain: "bridge.test", Tag: "bridge"}, bridgeTestDispatcher{})
	if err != nil {
		t.Fatal(err)
	}
	manager.workers = []*BridgeWorker{pair.bridge}
	t.Cleanup(func() {
		manager.Close()
		for _, w := range manager.workers {
			w.worker.Close()
		}
	})

	// Drop only Portal-to-Bridge frames; leave both pipes open.
	pair.requests.discard.Store(true)
	select {
	case <-pair.requests.dropped:
	case <-time.After(3 * controlHeartbeatInterval):
		t.Fatal("no control heartbeat was dropped")
	}
	pair.bridge.controlAccess.Lock()
	lastControl, state := pair.bridge.lastControl, pair.bridge.state
	pair.bridge.controlAccess.Unlock()
	if lastControl.IsZero() || state != Control_ACTIVE || pair.server.Closed() {
		t.Fatal("old bridge worker was not active when the one-way stall began")
	}
	deadline := time.Now().Add(controlSilenceTimeout + 6*time.Second)
	ticker := time.NewTicker(controlHeartbeatInterval)
	defer ticker.Stop()
	for time.Now().Before(deadline) {
		if err := manager.monitor(); err != nil {
			t.Fatal(err)
		}
		if len(manager.workers) == 1 && manager.workers[0] != pair.bridge && pair.server.Closed() {
			if elapsed := time.Since(lastControl); elapsed < controlSilenceTimeout {
				t.Fatalf("bridge replaced worker before heartbeat deadline: %v", elapsed)
			}
			if state := bridgeState(pair.bridge); state != Control_DRAIN {
				t.Fatalf("bridge did not retire the expired worker: state=%v", state)
			}
			return
		}
		<-ticker.C
	}
	t.Fatal("bridge did not replace its half-open worker")
}

// A legacy DRAIN changes state without opting into the heartbeat deadline.
func TestBridgeLegacyDrainDoesNotOptIn(t *testing.T) {
	bridge := &BridgeWorker{}
	control := &Control{State: Control_DRAIN, Random: []byte("legacy")}
	payload, err := proto.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	reader, writer := pipe.New(pipe.WithoutSizeLimit())
	replyReader, replyWriter := pipe.New(pipe.WithoutSizeLimit())
	bridge.handleInternalConn(transport.Link{Reader: reader, Writer: replyWriter})
	t.Cleanup(func() { reader.Interrupt(); replyReader.Interrupt() })
	if err := writer.WriteMultiBuffer(buf.MergeBytes(nil, payload)); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for bridgeState(bridge) != Control_DRAIN {
		if time.Now().After(deadline) {
			t.Fatal("legacy DRAIN state was ignored")
		}
		time.Sleep(time.Millisecond)
	}
	bridge.controlAccess.Lock()
	lastControl := bridge.lastControl
	bridge.controlAccess.Unlock()
	if !lastControl.IsZero() {
		t.Fatal("legacy DRAIN opted into the heartbeat deadline")
	}
	mb, err := replyReader.ReadMultiBufferTimeout(100 * time.Millisecond)
	buf.ReleaseMulti(mb)
	if err != buf.ErrReadTimeout {
		t.Fatalf("legacy DRAIN received an unsolicited acknowledgement: %v", err)
	}
}

type bridgeTestDispatcher struct{}

func (bridgeTestDispatcher) Type() interface{} { return routing.DispatcherType() }
func (bridgeTestDispatcher) Start() error      { return nil }
func (bridgeTestDispatcher) Close() error      { return nil }
func (bridgeTestDispatcher) Dispatch(context.Context, net.Destination) (*transport.Link, error) {
	reader, _ := pipe.New(pipe.WithoutSizeLimit())
	return &transport.Link{Reader: reader, Writer: buf.Discard}, nil
}

func TestBridgeReplacesWorkerAfterMissingControl(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	bridge, err := NewBridge(ctx, &BridgeConfig{Domain: "bridge.test", Tag: "bridge"}, bridgeTestDispatcher{})
	if err != nil {
		t.Fatal(err)
	}
	input, _ := pipe.New(pipe.WithoutSizeLimit())
	old := &BridgeWorker{}
	oldServer, err := mux.NewServerWorker(ctx, old, &transport.Link{Reader: input, Writer: buf.Discard})
	if err != nil {
		t.Fatal(err)
	}
	old.worker = oldServer
	bridge.workers = []*BridgeWorker{old}
	t.Cleanup(func() {
		bridge.Close()
		for _, w := range bridge.workers {
			w.worker.Close()
		}
		oldServer.Close()
		cancel()
	})

	// Do not impose a new timeout on a bridge serving an older portal.
	if old.expireIfUnhealthy(time.Now().Add(time.Hour)) || !old.IsActive() {
		t.Fatal("legacy portal was treated as inactive")
	}
	old.controlAccess.Lock()
	old.lastControl = time.Now().Add(-controlSilenceTimeout + time.Second)
	old.controlAccess.Unlock()
	if old.expireIfUnhealthy(time.Now()) || !old.IsActive() {
		t.Fatal("opted-in portal was expired before the deadline")
	}
	old.controlAccess.Lock()
	old.lastControl = time.Now().Add(-controlSilenceTimeout - time.Second)
	old.controlAccess.Unlock()
	if err := bridge.monitor(); err != nil {
		t.Fatal(err)
	}
	if len(bridge.workers) != 1 || bridge.workers[0] == old || bridgeState(old) != Control_DRAIN {
		t.Fatal("stale bridge worker was not replaced")
	}
	deadline := time.Now().Add(3 * time.Second)
	for !oldServer.Closed() {
		if time.Now().After(deadline) {
			t.Fatal("stale bridge server worker did not close")
		}
		time.Sleep(time.Millisecond)
	}
}
