//go:build !confonly
// +build !confonly

package rrpitTransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	gonet "net"
	"os"
	"path/filepath"
	"sync"
	"time"

	piondtls "github.com/pion/dtls/v3"
	"github.com/xtaci/smux"

	commonerrors "github.com/v2fly/v2ray-core/v5/common/errors"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	transportdtls "github.com/v2fly/v2ray-core/v5/transport/internet/dtls"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/packetToStream"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rriptMonoDirectionSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSession"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitBidirectionalSessionManager"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitChannelManager"
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitDebuger"
)

const (
	defaultDTLSMTU           = 1200
	defaultDTLSReplayWindow  = 128
	transportMaxPacketSize   = 8 << 20
	transportIdentityVersion = byte(1)
)

var (
	transportIdentityMagic  = [4]byte{'R', 'R', 'P', 'T'}
	transportHandshakeMagic = transportIdentityMagic
)

type errPathObjHolder struct{}

func newError(values ...interface{}) *commonerrors.Error {
	return commonerrors.New(values...).WithPathObj(errPathObjHolder{})
}

type transportSessionID [16]byte

type resolvedChannel struct {
	transport *ChannelSetting
	dtls      *DTLSUDPChannel
	address   v2net.Address
	port      v2net.Port
}

type ownedConn struct {
	gonet.Conn
	owner     *transportSession
	onClose   func()
	closeOnce sync.Once
	ctx       context.Context
	done      chan struct{}
}

type transportChannelSlot struct {
	config              rriptMonoDirectionSession.ChannelConfig
	conn                gonet.Conn
	managerChannelIndex int
	readIdleTimeout     time.Duration
}

type transportSession struct {
	id                     transportSessionID
	role                   string
	localSessionInstanceID rriptMonoDirectionSession.SessionInstanceID
	session                *rrpitBidirectionalSession.BidirectionalSession
	adaptor                *packetToStream.Adaptor

	sessionManager    *rrpitBidirectionalSessionManager.Manager
	backgroundAdaptor *packetToStream.Adaptor
	recorder          *rrpitDebuger.PacketRecorder
	onClose           func()
	persistence       connectionPersistencePolicy

	mu                sync.Mutex
	slots             []transportChannelSlot
	localAddr         gonet.Addr
	remoteAddr        gonet.Addr
	disconnected      bool
	disconnectedTimer *time.Timer

	closeOnce sync.Once
	closeErr  error
	closed    bool
}

func newTransportSession(
	role string,
	id transportSessionID,
	config *Config,
	client bool,
	onClose func(),
	onRemoteSessionInstance func(rriptMonoDirectionSession.SessionInstanceID) error,
) (*transportSession, error) {
	persistence := buildConnectionPersistencePolicy(config)
	channelManager, err := rrpitChannelManager.New(buildChannelManagerConfig(config))
	if err != nil {
		return nil, err
	}
	channelManager.SetBlockOnNoChannels(persistence.DisconnectedSessionRetention > 0)

	recorder, err := newPacketRecorder(config, role, id)
	if err != nil {
		_ = channelManager.Close()
		return nil, err
	}

	localSessionInstanceID, err := newSessionInstanceID()
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = channelManager.Close()
		return nil, err
	}

	sessionManagerConfig := buildBidirectionalSessionManagerConfig(config, channelManager)
	sessionManagerConfig.BaseSessionConfig.LocalSessionInstanceID = localSessionInstanceID
	if onRemoteSessionInstance != nil {
		sessionManagerConfig.BaseSessionConfig.ValidateRemoteControl = func(ctrl rriptMonoDirectionSession.ControlMessage) error {
			return onRemoteSessionInstance(ctrl.Session.InstanceID)
		}
	}

	sessionManager, err := rrpitBidirectionalSessionManager.New(sessionManagerConfig)
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = channelManager.Close()
		return nil, err
	}
	session := sessionManager.Session(rrpitBidirectionalSessionManager.InteractiveStream)
	backgroundSession := sessionManager.Session(rrpitBidirectionalSessionManager.BackgroundStream)

	maxMessageSize, err := session.MaxMessageSize()
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = sessionManager.Close()
		return nil, err
	}

	smuxConfig, err := buildSmuxConfig(config.GetAdaptor(), maxMessageSize)
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = sessionManager.Close()
		return nil, err
	}

	adaptor, err := packetToStream.New(session, client, smuxConfig)
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = sessionManager.Close()
		return nil, err
	}
	backgroundAdaptor, err := packetToStream.New(backgroundSession, client, smuxConfig)
	if err != nil {
		_ = adaptor.Close()
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = sessionManager.Close()
		return nil, err
	}

	return &transportSession{
		id:                     id,
		role:                   role,
		localSessionInstanceID: localSessionInstanceID,
		session:                session,
		adaptor:                adaptor,
		sessionManager:         sessionManager,
		backgroundAdaptor:      backgroundAdaptor,
		recorder:               recorder,
		onClose:                onClose,
		persistence:            persistence,
		slots:                  make([]transportChannelSlot, len(config.GetChannels())),
	}, nil
}

func (s *transportSession) attachChannel(conn gonet.Conn, channelSlot int, config rriptMonoDirectionSession.ChannelConfig, readIdleTimeout time.Duration) error {
	if s == nil || s.sessionManager == nil {
		return io.ErrClosedPipe
	}

	writer := io.WriteCloser(conn)
	if s.recorder != nil {
		writer = s.recorder.WrapWriter(s.role, channelSlot, writer)
	}

	managerChannelIndex, err := s.sessionManager.ChannelManager().AttachChannelWithConfig(writer, config)
	if err != nil {
		_ = conn.Close()
		return err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = s.sessionManager.ChannelManager().DetachChannel(managerChannelIndex)
		_ = conn.Close()
		return io.ErrClosedPipe
	}
	for len(s.slots) <= channelSlot {
		s.slots = append(s.slots, transportChannelSlot{})
	}
	oldConn := s.slots[channelSlot].conn
	oldManagerChannelIndex := s.slots[channelSlot].managerChannelIndex
	s.slots[channelSlot] = transportChannelSlot{
		config:              config,
		conn:                conn,
		managerChannelIndex: managerChannelIndex,
		readIdleTimeout:     readIdleTimeout,
	}
	if s.localAddr == nil {
		s.localAddr = conn.LocalAddr()
	}
	if s.remoteAddr == nil {
		s.remoteAddr = conn.RemoteAddr()
	}
	s.cancelDisconnectedTimerLocked()
	s.mu.Unlock()

	if oldConn != nil {
		_ = oldConn.Close()
		_ = s.sessionManager.ChannelManager().DetachChannel(oldManagerChannelIndex)
	}
	go s.readChannel(conn, channelSlot, managerChannelIndex, readIdleTimeout)
	return nil
}

func (s *transportSession) readChannel(conn gonet.Conn, channelSlot int, managerChannelIndex int, readIdleTimeout time.Duration) {
	packetBuf := make([]byte, transportMaxPacketSize)
	for {
		payload, err := readChannelPacket(conn, readIdleTimeout, packetBuf)
		if err != nil {
			s.handleChannelReadFailure(channelSlot, managerChannelIndex, conn, channelReadFailureMessage("failed to read channel packet", err), err)
			return
		}

		if s.recorder != nil {
			s.recorder.RecordInbound(s.role, channelSlot, payload)
		}
		if err := s.sessionManager.ChannelManager().OnNewMessageArrived(managerChannelIndex, payload); err != nil {
			if !s.isCurrentChannel(channelSlot, managerChannelIndex, conn) {
				_ = conn.Close()
				return
			}
			s.logChannelFailure(channelSlot, conn, "failed to handle inbound rrpit packet", err)
			_ = s.Close()
			return
		}
	}
}

func (s *transportSession) logChannelFailure(channelIndex int, conn gonet.Conn, message string, err error) {
	if s == nil || err == nil {
		return
	}

	var localAddr, remoteAddr interface{}
	if conn != nil {
		localAddr = conn.LocalAddr()
		remoteAddr = conn.RemoteAddr()
	}

	newError(
		"rrpit ", s.role, " channel ", channelIndex, ": ", message,
		", local=", localAddr, ", remote=", remoteAddr,
	).Base(err).AtWarning().WriteToLog()
}

func (s *transportSession) isCurrentChannel(channelSlot int, managerChannelIndex int, conn gonet.Conn) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isCurrentChannelLocked(channelSlot, managerChannelIndex, conn)
}

func (s *transportSession) isCurrentChannelLocked(channelSlot int, managerChannelIndex int, conn gonet.Conn) bool {
	if s.closed || channelSlot < 0 || channelSlot >= len(s.slots) {
		return false
	}
	slot := s.slots[channelSlot]
	return slot.conn == conn && slot.managerChannelIndex == managerChannelIndex
}

func (s *transportSession) handleChannelReadFailure(channelSlot int, managerChannelIndex int, conn gonet.Conn, message string, err error) {
	if s == nil {
		return
	}
	s.logChannelFailure(channelSlot, conn, message, err)

	s.mu.Lock()
	if !s.isCurrentChannelLocked(channelSlot, managerChannelIndex, conn) {
		s.mu.Unlock()
		_ = conn.Close()
		return
	}
	s.slots[channelSlot].conn = nil
	s.slots[channelSlot].managerChannelIndex = 0
	closeImmediately := s.activeChannelCountLocked() == 0 && s.scheduleDisconnectedTimerLocked()
	s.mu.Unlock()

	_ = s.sessionManager.ChannelManager().DetachChannel(managerChannelIndex)
	_ = conn.Close()
	if closeImmediately {
		_ = s.Close()
	}
}

func (s *transportSession) activeChannelCountLocked() int {
	count := 0
	for _, slot := range s.slots {
		if slot.conn != nil {
			count += 1
		}
	}
	return count
}

func (s *transportSession) cancelDisconnectedTimerLocked() {
	s.disconnected = false
	if s.disconnectedTimer != nil {
		s.disconnectedTimer.Stop()
		s.disconnectedTimer = nil
	}
}

func (s *transportSession) scheduleDisconnectedTimerLocked() bool {
	s.disconnected = true
	if s.disconnectedTimer != nil {
		s.disconnectedTimer.Stop()
		s.disconnectedTimer = nil
	}
	if s.persistence.DisconnectedSessionRetention <= 0 {
		return true
	}
	s.disconnectedTimer = time.AfterFunc(s.persistence.DisconnectedSessionRetention, func() {
		_ = s.Close()
	})
	return false
}

func (s *transportSession) MissingChannelSlots() []int {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	missing := make([]int, 0, len(s.slots))
	for index, slot := range s.slots {
		if slot.conn == nil {
			missing = append(missing, index)
		}
	}
	return missing
}

func (s *transportSession) Close() error {
	if s == nil {
		return nil
	}

	s.closeOnce.Do(func() {
		var channels []gonet.Conn
		s.mu.Lock()
		s.closed = true
		s.cancelDisconnectedTimerLocked()
		for index := range s.slots {
			if s.slots[index].conn != nil {
				channels = append(channels, s.slots[index].conn)
				s.slots[index].conn = nil
			}
		}
		s.mu.Unlock()

		if s.recorder != nil {
			s.closeErr = firstNonNil(s.closeErr, s.writeDiagnose())
		}
		for _, conn := range channels {
			s.closeErr = firstNonNil(s.closeErr, conn.Close())
		}
		// Stop the session manager before closing adaptors so its auto-tick loop
		// can't hold the bidirectional session lock while adaptor shutdown calls
		// back into session.Close().
		if s.sessionManager != nil {
			s.closeErr = firstNonNil(s.closeErr, s.sessionManager.Close())
		} else if s.session != nil {
			s.closeErr = firstNonNil(s.closeErr, s.session.Close())
		}
		if s.adaptor != nil {
			s.closeErr = firstNonNil(s.closeErr, s.adaptor.Close())
		}
		if s.backgroundAdaptor != nil {
			s.closeErr = firstNonNil(s.closeErr, s.backgroundAdaptor.Close())
		}

		if s.recorder != nil {
			s.closeErr = firstNonNil(s.closeErr, s.recorder.Close())
		}
		if s.onClose != nil {
			s.onClose()
		}
	})

	return s.closeErr
}

func (s *transportSession) writeDiagnose() error {
	if s == nil || s.recorder == nil || s.session == nil {
		return nil
	}
	manifest := s.recorder.Manifest()
	if manifest.Directory == "" {
		return nil
	}

	file, err := os.Create(filepath.Join(manifest.Directory, "session.json"))
	if err != nil {
		return err
	}
	defer file.Close()
	return rrpitDebuger.WriteDiagnoseOutput(file, s.session, s.recorder)
}

func (s *transportSession) OpenStream() (gonet.Conn, error) {
	return s.OpenStreamByClass(rrpitBidirectionalSessionManager.InteractiveStream)
}

func (s *transportSession) OpenStreamByClass(sessionClass rrpitBidirectionalSessionManager.SessionName) (gonet.Conn, error) {
	if s == nil || s.adaptor == nil {
		return nil, io.ErrClosedPipe
	}
	adaptor := s.adaptor
	if sessionClass == rrpitBidirectionalSessionManager.BackgroundStream {
		newError("opening background stream connection").AtDebug().WriteToLog()
		adaptor = s.backgroundAdaptor
	}
	if adaptor == nil {
		return nil, io.ErrClosedPipe
	}
	stream, err := adaptor.OpenStream()
	if err != nil {
		return nil, err
	}
	return gonet.Conn(stream), nil
}

func (s *transportSession) AcceptStream(sessionClass rrpitBidirectionalSessionManager.SessionName) (gonet.Conn, error) {
	if s == nil {
		return nil, io.ErrClosedPipe
	}
	adaptor := s.adaptor
	if sessionClass == rrpitBidirectionalSessionManager.BackgroundStream {
		adaptor = s.backgroundAdaptor
	}
	if adaptor == nil {
		return nil, io.ErrClosedPipe
	}
	stream, err := adaptor.AcceptStream()
	if err != nil {
		return nil, err
	}
	return gonet.Conn(stream), nil
}

func (c *ownedConn) GetConnectionContext() context.Context {
	if c == nil {
		return nil
	}
	return c.ctx
}

func (c *ownedConn) Close() error {
	if c == nil {
		return nil
	}
	defer c.closeOnce.Do(func() {
		if c.onClose != nil {
			c.onClose()
		}
		if c.done != nil {
			close(c.done)
		}
	})
	var err error
	if c.Conn != nil {
		err = c.Conn.Close()
	} else if c.owner != nil {
		err = c.owner.Close()
	}
	return err
}

func (c *ownedConn) LocalAddr() gonet.Addr {
	if c == nil {
		return nil
	}
	if c.owner != nil {
		c.owner.mu.Lock()
		addr := c.owner.localAddr
		c.owner.mu.Unlock()
		if addr != nil {
			return addr
		}
	}
	if c.Conn != nil {
		return c.Conn.LocalAddr()
	}
	return nil
}

func (c *ownedConn) RemoteAddr() gonet.Addr {
	if c == nil {
		return nil
	}
	if c.owner != nil {
		c.owner.mu.Lock()
		addr := c.owner.remoteAddr
		c.owner.mu.Unlock()
		if addr != nil {
			return addr
		}
	}
	if c.Conn != nil {
		return c.Conn.RemoteAddr()
	}
	return nil
}

func buildBidirectionalSessionConfig(config *Config) rrpitBidirectionalSession.Config {
	var lane *LaneSetting
	if config != nil {
		lane = config.GetLane()
	}
	var session *SessionSetting
	if config != nil {
		session = config.GetSession()
	}
	var reconstruction *SessionReconstructionSetting
	if session != nil {
		reconstruction = session.GetReconstruction()
	}

	return rrpitBidirectionalSession.Config{
		Rx: rriptMonoDirectionSession.SessionRxConfig{
			LaneShardSize:              int(lane.GetShardSize()),
			MaxBufferedLanes:           int(lane.GetMaxBufferedLanes()),
			RemoteMaxDataShardsPerLane: int(lane.GetRemoteMaxDataShardsPerLane()),
			OnMessage: func([]byte) error {
				return nil
			},
		},
		Tx: rriptMonoDirectionSession.SessionTxConfig{
			LaneShardSize:                  int(lane.GetShardSize()),
			MaxDataShardsPerLane:           int(lane.GetMaxDataShardsPerLane()),
			MaxBufferedLanes:               int(lane.GetMaxBufferedLanes()),
			MaxRewindableTimestampNum:      int(session.GetMaxRewindableTimestampNum()),
			MaxRewindableControlMessageNum: int(session.GetMaxRewindableControlMessageNum()),
			OddChannelIDs:                  session.GetOddChannelIds(),
			Reconstruction: rriptMonoDirectionSession.SessionTxReconstructionConfig{
				InitialRepairShardRatio:                       float64(reconstruction.GetInitialRepairShardRatio()),
				LaneRepairWeight:                              float32SliceToFloat64Slice(reconstruction.GetLaneRepairWeight()),
				SecondaryRepairShardRatio:                     float64(reconstruction.GetSecondaryRepairShardRatio()),
				TimeResendSecondaryRepairShard:                int(reconstruction.GetTimeResendSecondaryRepairShard()),
				StaleLaneFinalizedAgeThresholdTicks:           int(reconstruction.GetStaleLaneFinalizedAgeThresholdTicks()),
				StaleLaneProgressStallThresholdTicks:          int(reconstruction.GetStaleLaneProgressStallThresholdTicks()),
				SecondaryRepairMinBurst:                       int(reconstruction.GetSecondaryRepairMinBurst()),
				AlwaysRestrictSourceDataWhenOldestLaneStalled: reconstruction.GetAlwaysRestrictSourceDataWhenOldestLaneStalled(),
			},
		},
		TimestampInterval: 0,
	}
}

func buildChannelManagerConfig(config *Config) rrpitChannelManager.Config {
	var session *SessionSetting
	if config != nil {
		session = config.GetSession()
	}
	return rrpitChannelManager.Config{
		OddChannelIDs:                  session.GetOddChannelIds(),
		MaxRewindableTimestampNum:      int(session.GetMaxRewindableTimestampNum()),
		MaxRewindableControlMessageNum: int(session.GetMaxRewindableControlMessageNum()),
	}
}

func buildBidirectionalSessionManagerConfig(
	config *Config,
	channelManager *rrpitChannelManager.ChannelManager,
) rrpitBidirectionalSessionManager.Config {
	var sessionMgr *SessionManagerSetting
	if config != nil {
		sessionMgr = config.GetSessionMgr()
	}
	baseSessionConfig := buildBidirectionalSessionConfig(config)
	baseSessionConfig.ManagerHostedControlKeepaliveIntervalTicks = int(sessionMgr.GetManagerHostedControlKeepaliveIntervalTicks())
	return rrpitBidirectionalSessionManager.Config{
		ChannelManager:    channelManager,
		BaseSessionConfig: baseSessionConfig,
		TimestampInterval: time.Duration(sessionMgr.GetTimestampInterval()),
		InteractivePrimaryCancellationCounterLimit:          int(sessionMgr.GetInteractivePrimaryCancellationCounterLimit()),
		DynamicRestrictSourceDataWhenOldestLaneStalled:      sessionMgr.GetDynamicRestrictSourceDataWhenOldestLaneStalled(),
		DynamicRestrictSourceDataWhenOldestLaneStalledTicks: int(sessionMgr.GetDynamicRestrictSourceDataWhenOldestLaneStalledTicks()),
	}
}

func float32SliceToFloat64Slice(values []float32) []float64 {
	if len(values) == 0 {
		return nil
	}
	converted := make([]float64, len(values))
	for i, value := range values {
		converted[i] = float64(value)
	}
	return converted
}

func buildSmuxConfig(config *AdaptorSetting, maxMessageSize int) (*smux.Config, error) {
	smuxConfig := smux.DefaultConfig()
	if config == nil {
		if maxFrameSize := packetToStream.MaxSmuxFrameSizeForMessage(maxMessageSize); maxFrameSize > 0 && smuxConfig.MaxFrameSize > maxFrameSize {
			smuxConfig.MaxFrameSize = maxFrameSize
		}
		return smuxConfig, nil
	}

	if version := config.GetSmuxVersion(); version != 0 {
		smuxConfig.Version = int(version)
	}
	smuxConfig.KeepAliveDisabled = config.GetKeepAliveDisabled()
	if interval := config.GetKeepAliveInterval(); interval != 0 {
		smuxConfig.KeepAliveInterval = time.Duration(interval)
	}
	if timeout := config.GetKeepAliveTimeout(); timeout != 0 {
		smuxConfig.KeepAliveTimeout = time.Duration(timeout)
	}
	if frameSize := config.GetMaxFrameSize(); frameSize != 0 {
		smuxConfig.MaxFrameSize = int(frameSize)
	} else if maxFrameSize := packetToStream.MaxSmuxFrameSizeForMessage(maxMessageSize); maxFrameSize > 0 && smuxConfig.MaxFrameSize > maxFrameSize {
		smuxConfig.MaxFrameSize = maxFrameSize
	}
	if receiveBuffer := config.GetMaxReceiveBuffer(); receiveBuffer != 0 {
		smuxConfig.MaxReceiveBuffer = int(receiveBuffer)
	}
	if streamBuffer := config.GetMaxStreamBuffer(); streamBuffer != 0 {
		smuxConfig.MaxStreamBuffer = int(streamBuffer)
	}

	if err := smux.VerifyConfig(smuxConfig); err != nil {
		return nil, err
	}
	return smuxConfig, nil
}

func newPacketRecorder(config *Config, role string, id transportSessionID) (*rrpitDebuger.PacketRecorder, error) {
	if config == nil || config.GetEngineering() == nil || config.GetEngineering().GetDebuggerDir() == "" {
		return nil, nil
	}

	return rrpitDebuger.NewPacketRecorder(rrpitDebuger.PacketRecorderConfig{
		Directory:        filepath.Join(config.GetEngineering().GetDebuggerDir(), role+"-"+hex.EncodeToString(id[:])),
		MaxFileSizeBytes: config.GetEngineering().GetDebuggerMaxFileSizeBytes(),
		MaxFiles:         int(config.GetEngineering().GetDebuggerMaxFiles()),
	})
}

func resolveDialChannels(dest v2net.Destination, config *Config) ([]resolvedChannel, error) {
	return resolveChannels(func(index int, channel *DTLSUDPChannel) (v2net.Address, v2net.Port, error) {
		address := dest.Address
		port := dest.Port
		if resolved := channelAddress(channel); resolved != nil {
			address = resolved
		}
		if channel.GetPort() != 0 {
			port = v2net.Port(channel.GetPort())
		}
		if address == nil {
			return nil, 0, fmt.Errorf("rrpit channel %d missing remote address", index)
		}
		if port == 0 {
			return nil, 0, fmt.Errorf("rrpit channel %d missing remote port", index)
		}
		return address, port, nil
	}, config)
}

func resolveListenChannels(address v2net.Address, port v2net.Port, config *Config) ([]resolvedChannel, error) {
	return resolveChannels(func(index int, channel *DTLSUDPChannel) (v2net.Address, v2net.Port, error) {
		resolvedAddress := address
		resolvedPort := port
		if resolved := channelAddress(channel); resolved != nil {
			resolvedAddress = resolved
		}
		if channel.GetPort() != 0 {
			resolvedPort = v2net.Port(channel.GetPort())
		}
		if resolvedAddress == nil {
			return nil, 0, fmt.Errorf("rrpit channel %d missing listen address", index)
		}
		if resolvedPort == 0 {
			return nil, 0, fmt.Errorf("rrpit channel %d missing listen port", index)
		}
		return resolvedAddress, resolvedPort, nil
	}, config)
}

func resolveChannels(
	resolve func(index int, channel *DTLSUDPChannel) (v2net.Address, v2net.Port, error),
	config *Config,
) ([]resolvedChannel, error) {
	if config == nil || len(config.GetChannels()) == 0 {
		return nil, fmt.Errorf("rrpit requires at least one channel")
	}

	resolved := make([]resolvedChannel, 0, len(config.GetChannels()))
	for index, channel := range config.GetChannels() {
		if channel == nil || channel.GetChannel() == nil {
			return nil, fmt.Errorf("rrpit channel %d missing channel settings", index)
		}

		channelSetting := channel.GetSetting()
		if channelSetting == nil {
			channelSetting = &ChannelSetting{}
		}

		message, err := serial.GetInstanceOf(channel.GetChannel())
		if err != nil {
			return nil, err
		}

		dtlsChannel, ok := message.(*DTLSUDPChannel)
		if !ok {
			return nil, fmt.Errorf("rrpit channel %d uses unsupported transport %T", index, message)
		}

		address, resolvedPort, err := resolve(index, dtlsChannel)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, resolvedChannel{
			transport: channelSetting,
			dtls:      dtlsChannel,
			address:   address,
			port:      resolvedPort,
		})
	}
	return resolved, nil
}

func channelAddress(channel *DTLSUDPChannel) v2net.Address {
	if channel == nil {
		return nil
	}
	if len(channel.GetIp()) > 0 {
		return v2net.IPAddress(channel.GetIp())
	}
	if channel.GetIpAddr() != "" {
		return v2net.ParseAddress(channel.GetIpAddr())
	}
	return nil
}

func rrpitChannelConfig(config *ChannelSetting) rriptMonoDirectionSession.ChannelConfig {
	if config == nil {
		config = &ChannelSetting{}
	}
	return rriptMonoDirectionSession.ChannelConfig{
		Weight:          int(config.GetWeight()),
		MaxSendingSpeed: int(config.GetMaxSendingSpeed()),
	}
}

func channelReadIdleTimeout(config resolvedChannel) time.Duration {
	if config.dtls == nil {
		return 0
	}
	timeout := time.Duration(config.dtls.GetNoIncomingMessageTimeout())
	if timeout <= 0 {
		return 0
	}
	return timeout
}

func applyChannelReadDeadline(conn gonet.Conn, timeout time.Duration) error {
	if conn == nil {
		return io.ErrClosedPipe
	}
	if timeout <= 0 {
		return conn.SetReadDeadline(time.Time{})
	}
	return conn.SetReadDeadline(time.Now().Add(timeout))
}

func readChannelPacket(conn gonet.Conn, timeout time.Duration, buffer []byte) ([]byte, error) {
	if conn == nil {
		return nil, io.ErrClosedPipe
	}
	if len(buffer) == 0 {
		return nil, io.ErrShortBuffer
	}
	if err := applyChannelReadDeadline(conn, timeout); err != nil {
		return nil, err
	}
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buffer[:n]...), nil
}

func channelReadFailureMessage(base string, err error) string {
	if err == nil {
		return base
	}
	if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
		return "channel considered dead after no incoming message within configured timeout"
	}
	return base
}

func makeTransportDTLSConfig(config resolvedChannel) *transportdtls.Config {
	mtu := config.transport.GetMtu()
	if mtu == 0 {
		mtu = defaultDTLSMTU
	}
	return &transportdtls.Config{
		Mode:                   transportdtls.DTLSMode_PSK,
		Psk:                    []byte(config.dtls.GetPassword()),
		Mtu:                    mtu,
		ReplayProtectionWindow: defaultDTLSReplayWindow,
	}
}

func makePionDTLSConfig(config resolvedChannel, sessionID transportSessionID) *piondtls.Config {
	mtu := config.transport.GetMtu()
	if mtu == 0 {
		mtu = defaultDTLSMTU
	}
	return &piondtls.Config{
		MTU:                    int(mtu),
		ReplayProtectionWindow: defaultDTLSReplayWindow,
		PSK: func([]byte) ([]byte, error) {
			return []byte(config.dtls.GetPassword()), nil
		},
		PSKIdentityHint: encodeTransportSessionIdentity(sessionID),
		CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256},
	}
}

func newSessionID() (transportSessionID, error) {
	var id transportSessionID
	_, err := rand.Read(id[:])
	return id, err
}

func newSessionInstanceID() (rriptMonoDirectionSession.SessionInstanceID, error) {
	var id rriptMonoDirectionSession.SessionInstanceID
	_, err := rand.Read(id[:])
	return id, err
}

func encodeTransportSessionIdentity(id transportSessionID) []byte {
	identity := make([]byte, 0, len(transportIdentityMagic)+1+len(id))
	identity = append(identity, transportIdentityMagic[:]...)
	identity = append(identity, transportIdentityVersion)
	identity = append(identity, id[:]...)
	return identity
}

func readTransportSessionID(conn gonet.Conn) (transportSessionID, error) {
	var id transportSessionID

	identity := transportdtls.ClientIdentity(conn)
	if len(identity) == 0 {
		return id, fmt.Errorf("rrpit missing dtls client identity")
	}
	if len(identity) != len(transportIdentityMagic)+1+len(id) {
		return id, fmt.Errorf("rrpit invalid dtls client identity length")
	}
	for i, b := range transportIdentityMagic {
		if identity[i] != b {
			return id, fmt.Errorf("rrpit invalid dtls client identity magic")
		}
	}
	if identity[len(transportIdentityMagic)] != transportIdentityVersion {
		return id, fmt.Errorf("rrpit unsupported dtls client identity version")
	}
	copy(id[:], identity[len(transportIdentityMagic)+1:])
	return id, nil
}

func makeDTLSStreamSettings(config resolvedChannel, socketSettings *internet.SocketConfig) *internet.MemoryStreamConfig {
	return &internet.MemoryStreamConfig{
		ProtocolName:     "dtls",
		ProtocolSettings: makeTransportDTLSConfig(config),
		SocketSettings:   socketSettings,
	}
}

func firstNonNil(current error, next error) error {
	if current != nil {
		return current
	}
	return next
}
