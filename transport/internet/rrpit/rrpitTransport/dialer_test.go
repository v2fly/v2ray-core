package rrpitTransport

import (
	"errors"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
)

func TestPersistentClientSessionKeepsSessionOnMatchingRemoteInstanceID(t *testing.T) {
	closeCalled := make(chan struct{}, 1)
	session := &persistentClientSession{
		closeSession: func() error {
			closeCalled <- struct{}{}
			return nil
		},
	}

	remoteInstanceID := rriptMonoDirectionSession.SessionInstanceID{1, 2, 3}
	if err := session.handleRemoteSessionInstance(remoteInstanceID); err != nil {
		t.Fatal(err)
	}
	if err := session.handleRemoteSessionInstance(remoteInstanceID); err != nil {
		t.Fatal(err)
	}

	select {
	case <-closeCalled:
		t.Fatal("did not expect session close on matching remote instance id")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestPersistentClientSessionInvalidatesOnRemoteInstanceIDMismatch(t *testing.T) {
	state := &transportConnectionState{
		scopedSessionMap: map[string]*persistentClientSession{},
	}
	key := "rrpit-test"
	closeDone := make(chan struct{}, 1)

	session := &persistentClientSession{}
	session.closeSession = func() error {
		state.removeSession(key, session)
		closeDone <- struct{}{}
		return nil
	}
	state.scopedSessionMap[key] = session

	if err := session.handleRemoteSessionInstance(rriptMonoDirectionSession.SessionInstanceID{1}); err != nil {
		t.Fatal(err)
	}

	err := session.handleRemoteSessionInstance(rriptMonoDirectionSession.SessionInstanceID{2})
	if !errors.Is(err, errRemoteSessionRestarted) {
		t.Fatalf("unexpected mismatch error: %v", err)
	}

	select {
	case <-closeDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for stale session close")
	}

	if !session.IsClosed() {
		t.Fatal("expected session to be closed after remote instance mismatch")
	}
	state.scopedSessionAccess.Lock()
	_, ok := state.scopedSessionMap[key]
	state.scopedSessionAccess.Unlock()
	if ok {
		t.Fatal("expected stale session to be removed from cache")
	}
}

func TestPersistentClientSessionInvalidatesOnRemoteControlInactivity(t *testing.T) {
	state := &transportConnectionState{
		scopedSessionMap: map[string]*persistentClientSession{},
	}
	key := "rrpit-timeout"
	closeDone := make(chan struct{}, 1)

	session := &persistentClientSession{
		remoteControlInactivityTimeout: 20 * time.Millisecond,
	}
	session.closeSession = func() error {
		state.removeSession(key, session)
		closeDone <- struct{}{}
		return nil
	}
	state.scopedSessionMap[key] = session

	session.mu.Lock()
	session.armRemoteControlTimerLocked()
	session.mu.Unlock()

	select {
	case <-closeDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for remote control inactivity invalidation")
	}

	if !session.IsClosed() {
		t.Fatal("expected session to be closed after remote control inactivity")
	}
	state.scopedSessionAccess.Lock()
	_, ok := state.scopedSessionMap[key]
	state.scopedSessionAccess.Unlock()
	if ok {
		t.Fatal("expected stale session to be removed from cache after inactivity timeout")
	}
}
