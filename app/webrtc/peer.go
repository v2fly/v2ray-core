package webrtc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	pionwebrtc "github.com/pion/webrtc/v4"
)

const (
	dataChannelLabel = "v2ray-webrtc"
)

var (
	defaultDisconnectedTimeout = 5 * time.Second
	defaultFailedTimeout       = 25 * time.Second
	defaultKeepAliveInterval   = 2 * time.Second
)

type peerSession struct {
	pc *pionwebrtc.PeerConnection
	id string

	mu             sync.RWMutex
	dc             *pionwebrtc.DataChannel
	ready          chan struct{}
	done           chan struct{}
	once           sync.Once
	sendMu         sync.Mutex
	statsOnce      sync.Once
	connectedOnce  sync.Once
	gatherDoneOnce sync.Once
	err            error

	onCandidate   func([]byte)
	onConnected   func()
	onGatherDone  func()
	onDataChannel func(*peerDataChannel)
}

type peerDataChannel struct {
	owner *peerSession
	dc    *pionwebrtc.DataChannel
	label string

	ready chan struct{}
	done  chan struct{}
	once  sync.Once

	sendMu sync.Mutex

	handlerMu sync.RWMutex
	handler   func([]byte)

	err error
}

func newOffererPeer(api *pionwebrtc.API, cfg pionwebrtc.Configuration, id string, onCandidate func([]byte), onConnected func(), onGatherDone func(), onDataChannel func(*peerDataChannel)) (*peerSession, error) {
	peer, err := newPeerSession(api, cfg, id, onCandidate, onConnected, onGatherDone, onDataChannel)
	if err != nil {
		return nil, err
	}

	ordered := false
	maxRetransmits := uint16(0)
	dc, err := peer.pc.CreateDataChannel(dataChannelLabel, &pionwebrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	})
	if err != nil {
		peer.Close()
		return nil, newError("failed to create data channel").Base(err)
	}

	peer.bindControlDataChannel(dc)

	return peer, nil
}

func newAnswererPeer(api *pionwebrtc.API, cfg pionwebrtc.Configuration, id string, onCandidate func([]byte), onConnected func(), onGatherDone func(), onDataChannel func(*peerDataChannel)) (*peerSession, error) {
	peer, err := newPeerSession(api, cfg, id, onCandidate, onConnected, onGatherDone, onDataChannel)
	if err != nil {
		return nil, err
	}

	peer.pc.OnDataChannel(func(dc *pionwebrtc.DataChannel) {
		if dc == nil {
			return
		}
		if dc.Label() == dataChannelLabel {
			peer.bindControlDataChannel(dc)
			return
		}
		peer.bindApplicationDataChannel(dc, true)
	})

	return peer, nil
}

func newPeerSession(api *pionwebrtc.API, cfg pionwebrtc.Configuration, id string, onCandidate func([]byte), onConnected func(), onGatherDone func(), onDataChannel func(*peerDataChannel)) (*peerSession, error) {
	pc, err := api.NewPeerConnection(cfg)
	if err != nil {
		return nil, newError("failed to create peer connection").Base(err)
	}

	peer := &peerSession{
		id:            id,
		pc:            pc,
		ready:         make(chan struct{}),
		done:          make(chan struct{}),
		onCandidate:   onCandidate,
		onConnected:   onConnected,
		onGatherDone:  onGatherDone,
		onDataChannel: onDataChannel,
	}

	pc.OnICECandidate(func(candidate *pionwebrtc.ICECandidate) {
		if candidate == nil {
			peer.notifyGatherDone()
			return
		}
		if peer.onCandidate == nil {
			return
		}
		newError(
			"webrtc candidate local session=", peer.id,
			" candidate=", candidate.ToJSON().Candidate,
		).AtDebug().WriteToLog()
		raw, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			newError("failed to encode ICE candidate").Base(err).AtWarning().WriteToLog()
			return
		}
		peer.onCandidate(raw)
	})

	pc.OnSignalingStateChange(func(state pionwebrtc.SignalingState) {
		newError("webrtc signaling state session=", peer.id, " state=", state.String()).AtDebug().WriteToLog()
	})

	pc.OnICEGatheringStateChange(func(state pionwebrtc.ICEGatheringState) {
		newError("webrtc ice gathering state session=", peer.id, " state=", state.String()).AtDebug().WriteToLog()
		if state == pionwebrtc.ICEGatheringStateComplete {
			peer.notifyGatherDone()
		}
	})

	pc.OnICEConnectionStateChange(func(state pionwebrtc.ICEConnectionState) {
		newError("webrtc ice connection state session=", peer.id, " state=", state.String()).AtDebug().WriteToLog()
	})

	pc.OnConnectionStateChange(func(state pionwebrtc.PeerConnectionState) {
		newError("webrtc peer connection state session=", peer.id, " state=", state.String()).AtDebug().WriteToLog()
		if state == pionwebrtc.PeerConnectionStateConnected {
			peer.connectedOnce.Do(func() {
				if peer.onConnected != nil {
					peer.onConnected()
				}
				newError("webrtc connection successful session=", peer.id).AtDebug().WriteToLog()
				peer.logStatsSnapshot("connected")
			})
		}
		switch state {
		case pionwebrtc.PeerConnectionStateFailed:
			peer.closeWithError(newError("peer connection failed"))
		case pionwebrtc.PeerConnectionStateDisconnected:
			peer.closeWithError(newError("peer connection disconnected"))
		case pionwebrtc.PeerConnectionStateClosed:
			peer.closeWithError(nil)
		}
	})

	return peer, nil
}

func (p *peerSession) notifyGatherDone() {
	if p.onGatherDone == nil {
		return
	}
	p.gatherDoneOnce.Do(p.onGatherDone)
}

func (p *peerSession) bindControlDataChannel(dc *pionwebrtc.DataChannel) {
	p.mu.Lock()
	if p.dc != nil && p.dc != dc {
		p.mu.Unlock()
		return
	}
	p.dc = dc
	p.mu.Unlock()

	dc.OnOpen(func() {
		newError("webrtc control data channel open session=", p.id, " label=", dc.Label()).AtDebug().WriteToLog()
		p.onceReady()
		p.logStatsSnapshot("control-data-channel-open")
	})

	dc.OnClose(func() {
		newError("webrtc control data channel close session=", p.id, " label=", dc.Label()).AtDebug().WriteToLog()
	})

	dc.OnError(func(err error) {
		newError("webrtc control data channel error session=", p.id, " label=", dc.Label()).Base(err).AtDebug().WriteToLog()
	})
}

func (p *peerSession) onceReady() {
	select {
	case <-p.ready:
	default:
		newError("webrtc session ready session=", p.id).AtDebug().WriteToLog()
		close(p.ready)
	}
}

func (p *peerSession) bindApplicationDataChannel(dc *pionwebrtc.DataChannel, notify bool) *peerDataChannel {
	channel := &peerDataChannel{
		owner: p,
		dc:    dc,
		label: dc.Label(),
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}

	dc.OnOpen(func() {
		newError("webrtc data channel open session=", p.id, " label=", dc.Label()).AtDebug().WriteToLog()
		channel.onceReady()
	})

	dc.OnClose(func() {
		newError("webrtc data channel close session=", p.id, " label=", dc.Label()).AtDebug().WriteToLog()
		channel.closeWithError(nil)
	})

	dc.OnError(func(err error) {
		newError("webrtc data channel error session=", p.id, " label=", dc.Label()).Base(err).AtDebug().WriteToLog()
		channel.closeWithError(newError("data channel error").Base(err))
	})

	dc.OnMessage(func(msg pionwebrtc.DataChannelMessage) {
		channel.handlerMu.RLock()
		handler := channel.handler
		channel.handlerMu.RUnlock()
		if handler == nil {
			newError("dropping data channel message without handler session=", p.id, " label=", dc.Label()).AtDebug().WriteToLog()
			return
		}
		handler(append([]byte(nil), msg.Data...))
	})

	if notify && p.onDataChannel != nil {
		p.onDataChannel(channel)
	}

	return channel
}

func (p *peerSession) OpenDataChannel(label string) (*peerDataChannel, error) {
	select {
	case <-p.done:
		if p.err != nil {
			return nil, p.err
		}
		return nil, errors.New("peer session closed")
	default:
	}

	ordered := false
	maxRetransmits := uint16(0)
	dc, err := p.pc.CreateDataChannel(label, &pionwebrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	})
	if err != nil {
		return nil, newError("failed to create data channel").Base(err)
	}

	return p.bindApplicationDataChannel(dc, false), nil
}

func (p *peerSession) CreateOffer() ([]byte, error) {
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return nil, newError("failed to create WebRTC offer").Base(err)
	}
	if err := p.pc.SetLocalDescription(offer); err != nil {
		return nil, newError("failed to set local offer").Base(err)
	}
	local := p.pc.LocalDescription()
	if local == nil {
		return nil, newError("missing local offer description")
	}
	newError(
		"webrtc local offer session=", p.id,
		" sdp_candidates=", countSDPCandidates(local.SDP),
	).AtDebug().WriteToLog()
	return []byte(local.SDP), nil
}

func (p *peerSession) AcceptOffer(offer []byte) ([]byte, error) {
	newError(
		"webrtc accept remote offer session=", p.id,
		" sdp_candidates=", countSDPCandidates(string(offer)),
	).AtDebug().WriteToLog()
	if err := p.pc.SetRemoteDescription(pionwebrtc.SessionDescription{
		Type: pionwebrtc.SDPTypeOffer,
		SDP:  string(offer),
	}); err != nil {
		return nil, newError("failed to apply remote offer").Base(err)
	}

	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return nil, newError("failed to create WebRTC answer").Base(err)
	}
	if err := p.pc.SetLocalDescription(answer); err != nil {
		return nil, newError("failed to set local answer").Base(err)
	}
	local := p.pc.LocalDescription()
	if local == nil {
		return nil, newError("missing local answer description")
	}
	newError(
		"webrtc local answer session=", p.id,
		" sdp_candidates=", countSDPCandidates(local.SDP),
	).AtDebug().WriteToLog()
	return []byte(local.SDP), nil
}

func (p *peerSession) ApplyAnswer(answer []byte) error {
	if p.pc.RemoteDescription() != nil {
		return nil
	}
	newError(
		"webrtc apply remote answer session=", p.id,
		" sdp_candidates=", countSDPCandidates(string(answer)),
	).AtDebug().WriteToLog()
	return p.pc.SetRemoteDescription(pionwebrtc.SessionDescription{
		Type: pionwebrtc.SDPTypeAnswer,
		SDP:  string(answer),
	})
}

func (p *peerSession) AddICECandidate(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	var init pionwebrtc.ICECandidateInit
	if err := json.Unmarshal(raw, &init); err != nil {
		return newError("failed to decode ICE candidate").Base(err)
	}
	newError(
		"webrtc add remote candidate session=", p.id,
		" candidate=", init.Candidate,
	).AtDebug().WriteToLog()
	return p.pc.AddICECandidate(init)
}

func (p *peerSession) WaitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ready:
		return nil
	case <-p.done:
		if p.err != nil {
			return p.err
		}
		return errors.New("peer session closed before ready")
	}
}

func (p *peerSession) SendFrame(frame wireFrame) error {
	select {
	case <-p.done:
		if p.err != nil {
			return p.err
		}
		return errors.New("peer session closed")
	case <-p.ready:
	}

	data := encodeWireFrame(frame)
	if data == nil {
		return newError("failed to encode wire frame")
	}

	p.mu.RLock()
	dc := p.dc
	p.mu.RUnlock()
	if dc == nil {
		return newError("missing open data channel")
	}

	p.sendMu.Lock()
	defer p.sendMu.Unlock()

	return dc.Send(data)
}

func (p *peerSession) Done() <-chan struct{} {
	return p.done
}

func (c *peerDataChannel) Label() string {
	return c.label
}

func (c *peerDataChannel) SetMessageHandler(handler func([]byte)) {
	c.handlerMu.Lock()
	c.handler = handler
	c.handlerMu.Unlock()
}

func (c *peerDataChannel) WaitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ready:
		return nil
	case <-c.done:
		if c.err != nil {
			return c.err
		}
		return errors.New("data channel closed before ready")
	}
}

func (c *peerDataChannel) Send(payload []byte) error {
	select {
	case <-c.done:
		if c.err != nil {
			return c.err
		}
		return errors.New("data channel closed")
	case <-c.ready:
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	return c.dc.Send(append([]byte(nil), payload...))
}

func (c *peerDataChannel) Done() <-chan struct{} {
	return c.done
}

func (c *peerDataChannel) Close() error {
	c.closeWithError(nil)
	return nil
}

func (c *peerDataChannel) onceReady() {
	select {
	case <-c.ready:
	default:
		close(c.ready)
	}
}

func (c *peerDataChannel) closeWithError(err error) {
	c.once.Do(func() {
		c.err = err
		close(c.done)
		if c.dc != nil && c.dc.ReadyState() != pionwebrtc.DataChannelStateClosed {
			go func(dc *pionwebrtc.DataChannel) {
				_ = dc.Close()
			}(c.dc)
		}
	})
}

func (p *peerSession) Close() error {
	p.closeWithError(nil)
	return nil
}

func (p *peerSession) closeWithError(err error) {
	p.once.Do(func() {
		p.err = err
		if err != nil {
			newError("webrtc session close session=", p.id).Base(err).AtDebug().WriteToLog()
		} else {
			newError("webrtc session close session=", p.id, " clean=true").AtDebug().WriteToLog()
		}
		close(p.done)
		if p.pc != nil {
			go func(pc *pionwebrtc.PeerConnection) {
				_ = pc.Close()
			}(p.pc)
		}
	})
}

func (p *peerSession) logStatsSnapshot(reason string) {
	p.statsOnce.Do(func() {
		report := p.pc.GetStats()
		var pair *pionwebrtc.ICECandidatePairStats
		for _, stat := range report {
			candidatePair, ok := stat.(pionwebrtc.ICECandidatePairStats)
			if !ok {
				continue
			}
			if candidatePair.Nominated || candidatePair.State == pionwebrtc.StatsICECandidatePairStateSucceeded {
				cp := candidatePair
				pair = &cp
				break
			}
			if pair == nil {
				cp := candidatePair
				pair = &cp
			}
		}

		if pair == nil {
			newError("webrtc stats session=", p.id, " reason=", reason, " candidate_pair=none").AtDebug().WriteToLog()
			return
		}

		newError(
			"webrtc stats session=", p.id,
			" reason=", reason,
			" pair_state=", pair.State,
			" nominated=", pair.Nominated,
			" local_candidate_id=", pair.LocalCandidateID,
			" remote_candidate_id=", pair.RemoteCandidateID,
			" current_rtt=", pair.CurrentRoundTripTime,
			" bytes_sent=", pair.BytesSent,
			" bytes_recv=", pair.BytesReceived,
			" requests_sent=", pair.RequestsSent,
			" responses_recv=", pair.ResponsesReceived,
		).AtDebug().WriteToLog()
	})
}

func countSDPCandidates(sdp string) int {
	if sdp == "" {
		return 0
	}
	return strings.Count(sdp, "a=candidate:")
}
