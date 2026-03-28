package rrpitBidirectionalSessionManager

import (
	"io"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitChannelManager"
)

type SessionName string

const (
	InteractiveStream SessionName = "InteractiveStream"
	BackgroundStream  SessionName = "BackgroundStream"
)

type schedulerState int

const (
	interactiveStreamPrimary schedulerState = iota
	backgroundStreamPrimary
)

type Config struct {
	ChannelManager                                      *rrpitChannelManager.ChannelManager
	BaseSessionConfig                                   rrpitBidirectionalSession.Config
	TimestampInterval                                   time.Duration
	InteractivePrimaryCancellationCounterLimit          int
	DynamicRestrictSourceDataWhenOldestLaneStalled      bool
	DynamicRestrictSourceDataWhenOldestLaneStalledTicks int
}

type ManagerTickStats struct {
	InteractiveStream rriptMonoDirectionSession.TickStats
	BackgroundStream  rriptMonoDirectionSession.TickStats
}

type Manager struct {
	mu sync.Mutex

	channelManager *rrpitChannelManager.ChannelManager
	sessions       map[SessionName]*rrpitBidirectionalSession.BidirectionalSession

	state                                               schedulerState
	interactivePrimaryCancellationCounter               int
	interactivePrimaryCancellationLimit                 int
	dynamicRestrictSourceDataWhenOldestLaneStalled      bool
	dynamicRestrictSourceDataWhenOldestLaneStalledTicks int
	backgroundDynamicRestrictTicksRemaining             int
	timestampInterval                                   time.Duration
	nextTimestamp                                       uint64
	autoTickStop                                        chan struct{}
	autoTickDone                                        chan struct{}
	closeOnce                                           sync.Once
	closed                                              bool
}

func New(config Config) (*Manager, error) {
	if config.ChannelManager == nil {
		return nil, io.ErrClosedPipe
	}
	manager := &Manager{
		channelManager:                      config.ChannelManager,
		sessions:                            make(map[SessionName]*rrpitBidirectionalSession.BidirectionalSession),
		state:                               interactiveStreamPrimary,
		timestampInterval:                   config.TimestampInterval,
		interactivePrimaryCancellationLimit: config.InteractivePrimaryCancellationCounterLimit,
		dynamicRestrictSourceDataWhenOldestLaneStalled:      config.DynamicRestrictSourceDataWhenOldestLaneStalled,
		dynamicRestrictSourceDataWhenOldestLaneStalledTicks: config.DynamicRestrictSourceDataWhenOldestLaneStalledTicks,
	}

	interactiveConfig := config.BaseSessionConfig
	interactiveConfig.TimestampInterval = 0
	interactiveConfig.Rx.DataPacketKind = rriptMonoDirectionSession.PacketKind_InteractiveStreamData
	interactiveConfig.Rx.ControlPacketKind = rriptMonoDirectionSession.PacketKind_InteractiveStreamControl
	interactiveConfig.Tx.DataPacketKind = rriptMonoDirectionSession.PacketKind_InteractiveStreamData
	interactiveConfig.Tx.ControlPacketKind = rriptMonoDirectionSession.PacketKind_InteractiveStreamControl
	interactiveConfig.Tx.SendEnforced = config.ChannelManager.Send
	interactiveConfig.Tx.SendIgnoreQuota = config.ChannelManager.SendIgnoreQuota
	interactiveConfig.Tx.HasRemainingQuota = config.ChannelManager.HasRemainingQuota
	interactive, err := rrpitBidirectionalSession.New(interactiveConfig)
	if err != nil {
		return nil, err
	}

	backgroundConfig := config.BaseSessionConfig
	backgroundConfig.TimestampInterval = 0
	backgroundConfig.Rx.DataPacketKind = rriptMonoDirectionSession.PacketKind_BackgroundStreamData
	backgroundConfig.Rx.ControlPacketKind = rriptMonoDirectionSession.PacketKind_BackgroundStreamControl
	backgroundConfig.Tx.DataPacketKind = rriptMonoDirectionSession.PacketKind_BackgroundStreamData
	backgroundConfig.Tx.ControlPacketKind = rriptMonoDirectionSession.PacketKind_BackgroundStreamControl
	backgroundConfig.Tx.SendEnforced = config.ChannelManager.Send
	backgroundConfig.Tx.SendIgnoreQuota = config.ChannelManager.SendIgnoreQuota
	backgroundConfig.Tx.HasRemainingQuota = config.ChannelManager.HasRemainingQuota
	background, err := rrpitBidirectionalSession.New(backgroundConfig)
	if err != nil {
		_ = interactive.Close()
		return nil, err
	}

	manager.sessions[InteractiveStream] = interactive
	manager.sessions[BackgroundStream] = background
	config.ChannelManager.RegisterListener(rriptMonoDirectionSession.PacketKind_InteractiveStreamData, interactive.OnLogicalPacket)
	config.ChannelManager.RegisterListener(rriptMonoDirectionSession.PacketKind_InteractiveStreamControl, interactive.OnLogicalPacket)
	config.ChannelManager.RegisterListener(rriptMonoDirectionSession.PacketKind_BackgroundStreamData, background.OnLogicalPacket)
	config.ChannelManager.RegisterListener(rriptMonoDirectionSession.PacketKind_BackgroundStreamControl, background.OnLogicalPacket)

	manager.startAutoTick()
	return manager, nil
}

func (m *Manager) Session(name SessionName) *rrpitBidirectionalSession.BidirectionalSession {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[name]
}

func (m *Manager) ChannelManager() *rrpitChannelManager.ChannelManager {
	if m == nil {
		return nil
	}
	return m.channelManager
}

func (m *Manager) OnNewTimestamp(timestamp uint64) (ManagerTickStats, error) {
	if m == nil {
		return ManagerTickStats{}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ManagerTickStats{}, io.ErrClosedPipe
	}
	if timestamp > m.nextTimestamp {
		m.nextTimestamp = timestamp
	}
	m.channelManager.OnNewTimestamp(timestamp)
	interactiveSourceActivity := m.sessions[InteractiveStream].ConsumeSourcePacketActivity()
	backgroundDynamicRestrict := false
	if m.dynamicRestrictSourceDataWhenOldestLaneStalled {
		if interactiveSourceActivity > 0 {
			m.backgroundDynamicRestrictTicksRemaining = m.dynamicRestrictSourceDataWhenOldestLaneStalledTicks
			backgroundDynamicRestrict = true
		} else if m.backgroundDynamicRestrictTicksRemaining > 0 {
			m.backgroundDynamicRestrictTicksRemaining -= 1
			backgroundDynamicRestrict = true
		}
	}
	m.sessions[BackgroundStream].SetDynamicRestrictSourceDataWhenOldestLaneStalled(backgroundDynamicRestrict)

	var stats ManagerTickStats
	switch m.state {
	case interactiveStreamPrimary:
		interactive, err := m.sessions[InteractiveStream].OnNewTimestampWithStats(timestamp)
		stats.InteractiveStream = interactive
		if err != nil {
			return stats, err
		}
		if interactive.RepairPacketsGenerated > 0 {
			if m.interactivePrimaryCancellationCounter < m.interactivePrimaryCancellationLimit {
				m.interactivePrimaryCancellationCounter += 1
				m.state = interactiveStreamPrimary
				return stats, nil
			}
			m.interactivePrimaryCancellationCounter = 0
			m.state = backgroundStreamPrimary
			return stats, nil
		}
		if m.channelManager.HasRemainingQuota() {
			background, err := m.sessions[BackgroundStream].OnNewTimestampWithStats(timestamp)
			stats.BackgroundStream = background
			if err != nil {
				return stats, err
			}
		}
		m.interactivePrimaryCancellationCounter = 0
		m.state = backgroundStreamPrimary
		return stats, nil
	case backgroundStreamPrimary:
		background, err := m.sessions[BackgroundStream].OnNewTimestampWithStats(timestamp)
		stats.BackgroundStream = background
		if err != nil {
			return stats, err
		}
		if m.channelManager.HasRemainingQuota() {
			interactive, err := m.sessions[InteractiveStream].OnNewTimestampWithStats(timestamp)
			stats.InteractiveStream = interactive
			if err != nil {
				return stats, err
			}
		}
		m.state = interactiveStreamPrimary
		return stats, nil
	default:
		return stats, nil
	}
}

func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	var firstErr error
	m.closeOnce.Do(func() {
		// Wake any tick path blocked in the channel manager before trying to
		// take m.mu; OnNewTimestamp holds m.mu while sending control packets.
		if err := m.channelManager.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		m.mu.Lock()
		m.closed = true
		sessions := make([]*rrpitBidirectionalSession.BidirectionalSession, 0, len(m.sessions))
		for _, session := range m.sessions {
			sessions = append(sessions, session)
		}
		m.sessions = nil
		stop := m.autoTickStop
		done := m.autoTickDone
		m.mu.Unlock()

		if stop != nil {
			close(stop)
			if done != nil {
				<-done
			}
		}
		for _, session := range sessions {
			if err := session.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	})
	return firstErr
}

func (m *Manager) startAutoTick() {
	if m == nil || m.timestampInterval <= 0 {
		return
	}
	m.autoTickStop = make(chan struct{})
	m.autoTickDone = make(chan struct{})
	go func() {
		ticker := time.NewTicker(m.timestampInterval)
		defer func() {
			ticker.Stop()
			close(m.autoTickDone)
		}()
		for {
			select {
			case <-ticker.C:
				m.mu.Lock()
				m.nextTimestamp += 1
				timestamp := m.nextTimestamp
				m.mu.Unlock()
				if _, err := m.OnNewTimestamp(timestamp); err != nil {
					return
				}
			case <-m.autoTickStop:
				return
			}
		}
	}()
}
