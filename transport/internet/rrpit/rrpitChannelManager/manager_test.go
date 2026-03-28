package rrpitChannelManager

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
)

type recordingWriteCloser struct {
	mu     sync.Mutex
	writes [][]byte
}

func (w *recordingWriteCloser) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *recordingWriteCloser) Close() error { return nil }

func (w *recordingWriteCloser) snapshot() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([][]byte, len(w.writes))
	for i, write := range w.writes {
		out[i] = append([]byte(nil), write...)
	}
	return out
}

func TestChannelManagerFloodsControlAndStripsSharedFieldsOnReceive(t *testing.T) {
	manager, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	firstWriter := &recordingWriteCloser{}
	secondWriter := &recordingWriteCloser{}
	firstChannel, err := manager.AttachChannelWithConfig(firstWriter, rriptMonoDirectionSession.ChannelConfig{Weight: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AttachChannelWithConfig(secondWriter, rriptMonoDirectionSession.ChannelConfig{Weight: 1}); err != nil {
		t.Fatal(err)
	}

	var seen [][]byte
	manager.RegisterListener(rriptMonoDirectionSession.PacketKind_InteractiveStreamControl, func(payload []byte) error {
		seen = append(seen, append([]byte(nil), payload...))
		return nil
	})

	payload, err := rriptMonoDirectionSession.MarshalSessionControlPacket(
		rriptMonoDirectionSession.PacketKind_InteractiveStreamControl,
		rriptMonoDirectionSession.ControlMessage{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamControl, payload); err != nil {
		t.Fatal(err)
	}

	firstWrites := firstWriter.snapshot()
	secondWrites := secondWriter.snapshot()
	if len(firstWrites) != 1 || len(secondWrites) != 1 {
		t.Fatalf("expected one flooded control write per channel, got %d and %d", len(firstWrites), len(secondWrites))
	}

	firstCtrl, err := rriptMonoDirectionSession.UnmarshalSessionControlPacket(firstWrites[0][8:])
	if err != nil {
		t.Fatal(err)
	}
	secondCtrl, err := rriptMonoDirectionSession.UnmarshalSessionControlPacket(secondWrites[0][8:])
	if err != nil {
		t.Fatal(err)
	}
	if firstCtrl.FloodChannel.CurrentChannelID == 0 || secondCtrl.FloodChannel.CurrentChannelID == 0 {
		t.Fatal("expected flooded control to stamp non-zero channel ids")
	}
	if firstCtrl.FloodChannel.CurrentChannelID == secondCtrl.FloodChannel.CurrentChannelID {
		t.Fatal("expected flooded control clones to stamp distinct egress channel ids")
	}

	if err := manager.OnNewMessageArrived(firstChannel, firstWrites[0]); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 {
		t.Fatalf("expected one stripped control dispatch, got %d", len(seen))
	}
	strippedCtrl, err := rriptMonoDirectionSession.UnmarshalSessionControlPacket(seen[0])
	if err != nil {
		t.Fatal(err)
	}
	if strippedCtrl.FloodChannel.CurrentChannelID != 0 {
		t.Fatalf("expected stripped flood channel id 0, got %d", strippedCtrl.FloodChannel.CurrentChannelID)
	}
	if len(strippedCtrl.Channel.ChannelControl) != 0 || strippedCtrl.Channel.LenChannelControl != 0 {
		t.Fatalf("expected stripped channel control, got len field %d and %d entries", strippedCtrl.Channel.LenChannelControl, len(strippedCtrl.Channel.ChannelControl))
	}
}

func TestChannelManagerBypassSourceSelectionUsesLowestOversubscribe(t *testing.T) {
	manager, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	firstWriter := &recordingWriteCloser{}
	secondWriter := &recordingWriteCloser{}
	if _, err := manager.AttachChannelWithConfig(firstWriter, rriptMonoDirectionSession.ChannelConfig{Weight: 1, MaxSendingSpeed: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AttachChannelWithConfig(secondWriter, rriptMonoDirectionSession.ChannelConfig{Weight: 1, MaxSendingSpeed: 10}); err != nil {
		t.Fatal(err)
	}

	manager.OnNewTimestamp(1)
	if err := manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x01}); err != nil {
		t.Fatal(err)
	}
	if err := manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x02}); err != nil {
		t.Fatal(err)
	}

	if len(firstWriter.snapshot()) != 1 {
		t.Fatalf("expected first channel to carry only the first bypass source packet, got %d writes", len(firstWriter.snapshot()))
	}
	if len(secondWriter.snapshot()) != 1 {
		t.Fatalf("expected second channel to carry the second bypass source packet, got %d writes", len(secondWriter.snapshot()))
	}
}

func TestChannelManagerBypassTrafficDoesNotConsumeEnforcedQuota(t *testing.T) {
	manager, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	writer := &recordingWriteCloser{}
	if _, err := manager.AttachChannelWithConfig(writer, rriptMonoDirectionSession.ChannelConfig{Weight: 1, MaxSendingSpeed: 1}); err != nil {
		t.Fatal(err)
	}

	manager.OnNewTimestamp(1)
	if !manager.HasRemainingQuota() {
		t.Fatal("expected enforced quota to be available at start of tick")
	}
	if err := manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x01}); err != nil {
		t.Fatal(err)
	}
	if !manager.HasRemainingQuota() {
		t.Fatal("expected bypass traffic to leave enforced quota available")
	}
	if err := manager.Send(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x02}); err != nil {
		t.Fatalf("expected one enforced send after bypass traffic, got %v", err)
	}
	if manager.HasRemainingQuota() {
		t.Fatal("expected enforced quota to be exhausted after one enforced send")
	}
}

func TestChannelManagerDetachRemovesChannelFromSelection(t *testing.T) {
	manager, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	writer := &recordingWriteCloser{}
	channelIndex, err := manager.AttachChannelWithConfig(writer, rriptMonoDirectionSession.ChannelConfig{Weight: 1, MaxSendingSpeed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !manager.HasRemainingQuota() {
		t.Fatal("expected quota with attached channel")
	}
	if err := manager.DetachChannel(channelIndex); err != nil {
		t.Fatal(err)
	}
	if manager.HasRemainingQuota() {
		t.Fatal("expected no remaining quota after detaching the only channel")
	}
	if err := manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x01}); err == nil {
		t.Fatal("expected send to fail with no attached channels")
	} else if err != io.ErrClosedPipe {
		t.Fatalf("expected io.ErrClosedPipe after detach, got %v", err)
	}
}

func TestChannelManagerWaitsForChannelWhenConfigured(t *testing.T) {
	manager, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	manager.SetBlockOnNoChannels(true)

	done := make(chan error, 1)
	go func() {
		done <- manager.SendIgnoreQuota(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, []byte{rriptMonoDirectionSession.PacketKind_InteractiveStreamData, 0x01})
	}()

	select {
	case err := <-done:
		t.Fatalf("expected send to wait for a channel, got early result %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	writer := &recordingWriteCloser{}
	if _, err := manager.AttachChannelWithConfig(writer, rriptMonoDirectionSession.ChannelConfig{Weight: 1}); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for send to complete after channel attach")
	}
	if len(writer.snapshot()) != 1 {
		t.Fatalf("expected one write after attach, got %d", len(writer.snapshot()))
	}
}
