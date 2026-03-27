package rrpitBidirectionalSessionManager

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitChannelManager"
)

type recordingWriteCloser struct {
	writes [][]byte
}

func (w *recordingWriteCloser) Write(p []byte) (int, error) {
	w.writes = append(w.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (w *recordingWriteCloser) Close() error { return nil }

func TestManagerDynamicRestrictSourceDataWhenOldestLaneStalledTracksInteractiveActivity(t *testing.T) {
	channelManager, err := rrpitChannelManager.New(rrpitChannelManager.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer channelManager.Close()

	writer := &recordingWriteCloser{}
	if _, err := channelManager.AttachChannelWithConfig(writer, rriptMonoDirectionSession.ChannelConfig{Weight: 1, MaxSendingSpeed: 8}); err != nil {
		t.Fatal(err)
	}

	manager, err := New(Config{
		ChannelManager: channelManager,
		BaseSessionConfig: rrpitBidirectionalSession.Config{
			Rx: rriptMonoDirectionSession.SessionRxConfig{
				LaneShardSize:    16,
				MaxBufferedLanes: 8,
				OnMessage:        func([]byte) error { return nil },
			},
			Tx: rriptMonoDirectionSession.SessionTxConfig{
				LaneShardSize:                  16,
				MaxDataShardsPerLane:           1,
				MaxBufferedLanes:               8,
				MaxRewindableTimestampNum:      8,
				MaxRewindableControlMessageNum: 8,
				Reconstruction: rriptMonoDirectionSession.SessionTxReconstructionConfig{
					InitialRepairShardRatio:              1,
					SecondaryRepairShardRatio:            1,
					TimeResendSecondaryRepairShard:       1,
					StaleLaneFinalizedAgeThresholdTicks:  2,
					StaleLaneProgressStallThresholdTicks: 2,
				},
			},
		},
		DynamicRestrictSourceDataWhenOldestLaneStalled:      true,
		DynamicRestrictSourceDataWhenOldestLaneStalledTicks: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	background := manager.Session(BackgroundStream)
	if background.RestrictSourceDataWhenOldestLaneStalledEnabled() {
		t.Fatal("expected background restriction to start disabled")
	}

	if err := manager.Session(InteractiveStream).SendMessage([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.OnNewTimestamp(1); err != nil {
		t.Fatal(err)
	}
	if !background.RestrictSourceDataWhenOldestLaneStalledEnabled() {
		t.Fatal("expected interactive source activity to enable dynamic background restriction")
	}

	if _, err := manager.OnNewTimestamp(2); err != nil {
		t.Fatal(err)
	}
	if !background.RestrictSourceDataWhenOldestLaneStalledEnabled() {
		t.Fatal("expected background restriction to stay enabled during retention window")
	}

	if _, err := manager.OnNewTimestamp(3); err != nil {
		t.Fatal(err)
	}
	if !background.RestrictSourceDataWhenOldestLaneStalledEnabled() {
		t.Fatal("expected background restriction to stay enabled for the second retained tick")
	}

	if _, err := manager.OnNewTimestamp(4); err != nil {
		t.Fatal(err)
	}
	if background.RestrictSourceDataWhenOldestLaneStalledEnabled() {
		t.Fatal("expected background restriction to clear after retained ticks expire")
	}
}
