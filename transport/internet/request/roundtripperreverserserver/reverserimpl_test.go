package roundtripperreverserserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

// helper to make a byte slice of given length filled with a pattern
func fill(b byte, n int) []byte {
	res := make([]byte, n)
	for i := range res {
		res[i] = b
	}
	return res
}

// stopCleanup tries to call StopCleanup on concrete *ReverserImpl returned by NewReverserImpl.
func stopCleanup(t *testing.T, r request.ReverserImpl) {
	if r == nil {
		return
	}
	if impl, ok := r.(*ReverserImpl); ok {
		if err := impl.StopCleanup(); err != nil {
			t.Fatalf("StopCleanup failed: %v", err)
		}
	} else {
		t.Fatalf("expected *ReverserImpl, got %T", r)
	}
}

func TestOnOtherRoundTrip_InvalidRoutingKeyLength(t *testing.T) {
	r, err := NewReverserImpl()
	if err != nil {
		// constructor currently never errors
		t.Fatalf("unexpected constructor error: %v", err)
	}
	defer stopCleanup(t, r)
	_, gotErr := r.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: []byte("short")})
	if gotErr == nil {
		t.Fatalf("expected error for invalid routing key length")
	}
	if !strings.Contains(gotErr.Error(), "invalid routing key") {
		t.Fatalf("unexpected error: %v", gotErr)
	}
}

func TestOnOtherRoundTrip_ServerToClient_NoClientFound(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	source := fill('A', 16)
	dest := fill('B', 16)
	impl.serverPrivateKeyToPublicKey.Store(string(source), string(source))
	_, err := impl.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: append(source, dest...)})
	if err == nil || !strings.Contains(err.Error(), "no client found") {
		t.Fatalf("expected no client found error, got: %v", err)
	}
}

func TestOnOtherRoundTrip_ServerToClient_Success(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	source := fill('S', 16)
	dest := fill('C', 16)
	impl.serverPrivateKeyToPublicKey.Store(string(source), string(source))
	clientState := &clientState{messageQueue: make(chan *reverserMessage, 1)}
	impl.clientTemporaryKeyToStateMap.Store(string(dest), clientState)
	data := []byte("hello-client")
	_, err := impl.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: append(source, dest...), Data: data})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case msg := <-clientState.messageQueue:
		if string(msg.Data) != string(data) {
			t.Fatalf("unexpected message data: got %q want %q", msg.Data, data)
		}
	default:
		t.Fatalf("expected message queued for client but queue empty")
	}
}

func TestOnOtherRoundTrip_ServerToClient_QueueFull(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	source := fill('F', 16)
	dest := fill('Q', 16)
	impl.serverPrivateKeyToPublicKey.Store(string(source), string(source))
	clientState := &clientState{messageQueue: make(chan *reverserMessage, 1)}
	// pre-fill queue so next send fails non-blocking
	clientState.messageQueue <- &reverserMessage{Data: []byte("prefill")}
	impl.clientTemporaryKeyToStateMap.Store(string(dest), clientState)
	_, err := impl.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: append(source, dest...), Data: []byte("new")})
	if err == nil || !strings.Contains(err.Error(), "client message queue full") {
		t.Fatalf("expected queue full error, got: %v", err)
	}
}

func TestOnOtherRoundTrip_ClientToServer_NoServerFound(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	source := fill('X', 16)
	dest := fill('Y', 16)
	_, err := impl.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: append(source, dest...), Data: []byte("ping")})
	if err == nil || !strings.Contains(err.Error(), "no server found") {
		t.Fatalf("expected no server found error, got: %v", err)
	}
}

func TestOnOtherRoundTrip_ClientToServer_Success(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	source := fill('1', 16)
	dest := fill('2', 16)
	serverState := &serverState{messageQueue: make(chan *reverserMessage, 1)}
	impl.serverPublicKeyToStateMap.Store(string(dest), serverState)
	clientState := &clientState{messageQueue: make(chan *reverserMessage, 1)}
	clientState.messageQueue <- &reverserMessage{Data: []byte("pong")}
	impl.clientTemporaryKeyToStateMap.Store(string(source), clientState)
	data := []byte("ping")
	resp, err := impl.OnOtherRoundTrip(context.Background(), request.Request{ConnectionTag: append(source, dest...), Data: data})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// verify server received the original message
	select {
	case msg := <-serverState.messageQueue:
		if string(msg.Data) != string(data) {
			t.Fatalf("server message mismatch: got %q want %q", msg.Data, data)
		}
	default:
		t.Fatalf("expected server to receive message but queue empty")
	}
	if string(resp.Data) != "pong" {
		t.Fatalf("unexpected response data: got %q want %q", resp.Data, "pong")
	}
}

func TestOnAuthenticatedServerIntentRoundTrip_InvalidPrivateKeyLength(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	serverPublic := fill('P', 16)
	_, err := impl.OnAuthenticatedServerIntentRoundTrip(context.Background(), serverPublic, request.Request{ConnectionTag: fill('Z', 15)})
	if err == nil || !strings.Contains(err.Error(), "invalid server private key") {
		t.Fatalf("expected invalid server private key error, got: %v", err)
	}
}

func TestOnAuthenticatedServerIntentRoundTrip_InvalidPublicKeyLength(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	_, err := impl.OnAuthenticatedServerIntentRoundTrip(context.Background(), fill('P', 15), request.Request{ConnectionTag: fill('K', 16)})
	if err == nil || !strings.Contains(err.Error(), "invalid server public key") {
		t.Fatalf("expected invalid server public key error, got: %v", err)
	}
}

func TestOnAuthenticatedServerIntentRoundTrip_SuccessMessage(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)
	serverPublic := fill('P', 16)
	serverPrivate := fill('K', 16)
	state := &serverState{messageQueue: make(chan *reverserMessage, 1)}
	state.messageQueue <- &reverserMessage{Data: []byte("welcome")}
	impl.serverPublicKeyToStateMap.Store(string(serverPublic), state)
	resp, err := impl.OnAuthenticatedServerIntentRoundTrip(context.Background(), serverPublic, request.Request{ConnectionTag: serverPrivate})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Data) != "welcome" {
		t.Fatalf("unexpected response data: got %q want %q", resp.Data, "welcome")
	}
	if _, ok := impl.serverPrivateKeyToPublicKey.Load(string(serverPrivate)); !ok {
		t.Fatalf("expected server private key mapping to be stored")
	}
}

func TestSmoke(t *testing.T) {
	if _, err := NewReverserImpl(); err != nil {
		t.Fatalf("unexpected error constructing reverser: %v", err)
	}
}

func TestPeriodicCleanupRemovesOldServer(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)

	// configure short durations for test speed
	impl.timeoutDuration = 10 * time.Millisecond
	impl.periodicCleaner.Interval = 5 * time.Millisecond

	// insert a server state with lastSeen far in the past
	old := &serverState{messageQueue: make(chan *reverserMessage, 1), lastSeen: time.Now().Add(-time.Hour)}
	impl.serverPublicKeyToStateMap.Store("old-server", old)

	// trigger cleanup now
	impl.__CleanupNow____TestOnly()

	if _, ok := impl.serverPublicKeyToStateMap.Load("old-server"); ok {
		t.Fatalf("expected old server entry to be removed by cleaner")
	}
}

func TestCleanupNowRemovesServerAndPrivateMapping(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)

	// create an old server entry with private key present in private->public map
	publicKey := "old-server-public"
	privateKey := "old-server-private"
	old := &serverState{messageQueue: make(chan *reverserMessage, 1), lastSeen: time.Now().Add(-time.Hour), privateKey: privateKey}
	impl.serverPublicKeyToStateMap.Store(publicKey, old)
	impl.serverPrivateKeyToPublicKey.Store(privateKey, publicKey)

	// trigger immediate cleanup
	impl.__CleanupNow____TestOnly()

	if _, ok := impl.serverPublicKeyToStateMap.Load(publicKey); ok {
		t.Fatalf("expected old server entry to be removed by cleaner via CleanupNow")
	}
	if _, ok := impl.serverPrivateKeyToPublicKey.Load(privateKey); ok {
		t.Fatalf("expected old private->public mapping to be removed by cleaner via CleanupNow")
	}
}

func TestStopCleanupIdempotent(t *testing.T) {
	r, _ := NewReverserImpl()
	impl := r.(*ReverserImpl)
	// first stop
	if err := impl.StopCleanup(); err != nil {
		t.Fatalf("StopCleanup failed: %v", err)
	}
	// second stop should not panic or return error
	if err := impl.StopCleanup(); err != nil {
		t.Fatalf("StopCleanup second call failed: %v", err)
	}
}

func TestOnAuthenticatedServerIntentRoundTrip_UpdatesLastSeenAndMapping(t *testing.T) {
	r, _ := NewReverserImpl()
	defer stopCleanup(t, r)
	impl := r.(*ReverserImpl)

	serverPublic := fill('P', 16)
	serverPrivate := fill('K', 16)

	// pre-insert a server state that has an old lastSeen but contains a queued message
	state := &serverState{messageQueue: make(chan *reverserMessage, 1), lastSeen: time.Now().Add(-time.Hour)}
	state.messageQueue <- &reverserMessage{Data: []byte("welcome-back")}
	impl.serverPublicKeyToStateMap.Store(string(serverPublic), state)

	resp, err := impl.OnAuthenticatedServerIntentRoundTrip(context.Background(), serverPublic, request.Request{ConnectionTag: serverPrivate})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Data) != "welcome-back" {
		t.Fatalf("unexpected response data: got %q want %q", resp.Data, "welcome-back")
	}

	// mapping from private -> public should be stored
	if v, ok := impl.serverPrivateKeyToPublicKey.Load(string(serverPrivate)); !ok || v != string(serverPublic) {
		t.Fatalf("expected private->public mapping stored, got %v (ok=%v)", v, ok)
	}

	// lastSeen should be updated to a recent time
	si, ok := impl.serverPublicKeyToStateMap.Load(string(serverPublic))
	if !ok {
		t.Fatalf("expected server state to still exist after intent call")
	}
	ss := si.(*serverState)
	if time.Since(ss.lastSeen) > time.Second*2 {
		t.Fatalf("expected lastSeen to be updated recently, got %v", ss.lastSeen)
	}
}
