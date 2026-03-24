//go:build !confonly
// +build !confonly

package rrpitTransport

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
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
	"github.com/v2fly/v2ray-core/v5/transport/internet/rrpit/rrpitDebuger"
)

const (
	defaultDTLSMTU                = 1200
	defaultDTLSReplayWindow       = 128
	transportFrameLengthFieldSize = 4
	transportMaxFrameSize         = 8 << 20
	transportIdentityVersion      = byte(1)
)

var transportIdentityMagic = [4]byte{'R', 'R', 'P', 'T'}
var transportHandshakeMagic = transportIdentityMagic

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

type framedPacketWriteCloser struct {
	conn gonet.Conn
	mu   sync.Mutex
}

type ownedConn struct {
	gonet.Conn
	owner     *transportSession
	onClose   func()
	closeOnce sync.Once
	ctx       context.Context
	done      chan struct{}
}

type transportSession struct {
	id       transportSessionID
	role     string
	session  *rrpitBidirectionalSession.BidirectionalSession
	adaptor  *packetToStream.Adaptor
	recorder *rrpitDebuger.PacketRecorder
	onClose  func()

	mu         sync.Mutex
	channels   []gonet.Conn
	localAddr  gonet.Addr
	remoteAddr gonet.Addr

	closeOnce sync.Once
	closeErr  error
}

func newTransportSession(
	role string,
	id transportSessionID,
	config *Config,
	client bool,
	onClose func(),
) (*transportSession, error) {
	sessionConfig := buildBidirectionalSessionConfig(config)
	session, err := rrpitBidirectionalSession.New(sessionConfig)
	if err != nil {
		return nil, err
	}

	recorder, err := newPacketRecorder(config, role, id)
	if err != nil {
		_ = session.Close()
		return nil, err
	}

	smuxConfig, err := buildSmuxConfig(config.GetAdaptor())
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = session.Close()
		return nil, err
	}

	adaptor, err := packetToStream.New(session, client, smuxConfig)
	if err != nil {
		if recorder != nil {
			_ = recorder.Close()
		}
		_ = session.Close()
		return nil, err
	}

	return &transportSession{
		id:       id,
		role:     role,
		session:  session,
		adaptor:  adaptor,
		recorder: recorder,
		onClose:  onClose,
	}, nil
}

func (s *transportSession) addChannel(conn gonet.Conn, channelIndex int, config rriptMonoDirectionSession.ChannelConfig) error {
	if s == nil || s.session == nil {
		return io.ErrClosedPipe
	}

	writer := io.WriteCloser(&framedPacketWriteCloser{conn: conn})
	if s.recorder != nil {
		writer = s.recorder.WrapWriter(s.role, channelIndex, writer)
	}

	_, rxChannel, err := s.session.AttachChannelWithConfig(writer, config)
	if err != nil {
		_ = conn.Close()
		return err
	}

	s.mu.Lock()
	if s.localAddr == nil {
		s.localAddr = conn.LocalAddr()
	}
	if s.remoteAddr == nil {
		s.remoteAddr = conn.RemoteAddr()
	}
	s.channels = append(s.channels, conn)
	s.mu.Unlock()

	go s.readChannel(conn, rxChannel, channelIndex)
	return nil
}

func (s *transportSession) readChannel(conn gonet.Conn, rxChannel interface{ OnNewMessageArrived([]byte) error }, channelIndex int) {
	reader := bufio.NewReaderSize(conn, transportMaxFrameSize+transportFrameLengthFieldSize)
	for {
		lengthBuf := make([]byte, transportFrameLengthFieldSize)
		if _, err := io.ReadFull(reader, lengthBuf); err != nil {
			s.logChannelFailure(channelIndex, conn, "failed to read frame length", err)
			_ = s.Close()
			return
		}
		length := binary.BigEndian.Uint32(lengthBuf)
		if length == 0 || length > transportMaxFrameSize {
			s.logChannelFailure(channelIndex, conn, "received invalid frame length", fmt.Errorf("length=%d", length))
			_ = s.Close()
			return
		}

		payload := make([]byte, length)
		if _, err := io.ReadFull(reader, payload); err != nil {
			s.logChannelFailure(channelIndex, conn, "failed to read frame payload", err)
			_ = s.Close()
			return
		}

		if s.recorder != nil {
			s.recorder.RecordInbound(s.role, channelIndex, payload)
		}
		if err := rxChannel.OnNewMessageArrived(payload); err != nil {
			s.logChannelFailure(channelIndex, conn, "failed to handle inbound rrpit packet", err)
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

func (s *transportSession) Close() error {
	if s == nil {
		return nil
	}

	s.closeOnce.Do(func() {
		if s.recorder != nil {
			s.closeErr = firstNonNil(s.closeErr, s.writeDiagnose())
		}
		if s.adaptor != nil {
			s.closeErr = firstNonNil(s.closeErr, s.adaptor.Close())
		} else if s.session != nil {
			s.closeErr = firstNonNil(s.closeErr, s.session.Close())
		}

		s.mu.Lock()
		channels := append([]gonet.Conn(nil), s.channels...)
		s.channels = nil
		s.mu.Unlock()
		for _, conn := range channels {
			s.closeErr = firstNonNil(s.closeErr, conn.Close())
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
	if s == nil || s.adaptor == nil {
		return nil, io.ErrClosedPipe
	}
	stream, err := s.adaptor.OpenStream()
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

func (w *framedPacketWriteCloser) Write(p []byte) (int, error) {
	if w == nil || w.conn == nil {
		return 0, io.ErrClosedPipe
	}
	if len(p) > int(^uint32(0)) {
		return 0, fmt.Errorf("rrpit channel payload too large")
	}

	frame := make([]byte, transportFrameLengthFieldSize+len(p))
	binary.BigEndian.PutUint32(frame[:transportFrameLengthFieldSize], uint32(len(p)))
	copy(frame[transportFrameLengthFieldSize:], p)

	w.mu.Lock()
	defer w.mu.Unlock()
	if err := writeAll(w.conn, frame); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *framedPacketWriteCloser) Close() error {
	if w == nil || w.conn == nil {
		return nil
	}
	return w.conn.Close()
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
			LaneShardSize:    int(lane.GetShardSize()),
			MaxBufferedLanes: int(lane.GetMaxBufferedLanes()),
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
				InitialRepairShardRatio:              float64(reconstruction.GetInitialRepairShardRatio()),
				LaneRepairWeight:                     float32SliceToFloat64Slice(reconstruction.GetLaneRepairWeight()),
				SecondaryRepairShardRatio:            float64(reconstruction.GetSecondaryRepairShardRatio()),
				TimeResendSecondaryRepairShard:       int(reconstruction.GetTimeResendSecondaryRepairShard()),
				StaleLaneFinalizedAgeThresholdTicks:  int(reconstruction.GetStaleLaneFinalizedAgeThresholdTicks()),
				StaleLaneProgressStallThresholdTicks: int(reconstruction.GetStaleLaneProgressStallThresholdTicks()),
				SecondaryRepairMinBurst:              int(reconstruction.GetSecondaryRepairMinBurst()),
			},
		},
		TimestampInterval: time.Duration(session.GetTimestampInterval()),
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

func buildSmuxConfig(config *AdaptorSetting) (*smux.Config, error) {
	smuxConfig := smux.DefaultConfig()
	if config == nil {
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

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}
		if written <= 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func firstNonNil(current error, next error) error {
	if current != nil {
		return current
	}
	return next
}
