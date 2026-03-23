package rrpitTransferChannel

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type ringBuffer[T any] struct {
	values []T
	start  int
	size   int
}

func newRingBuffer[T any](capacity int) ringBuffer[T] {
	if capacity <= 0 {
		return ringBuffer[T]{}
	}
	return ringBuffer[T]{
		values: make([]T, capacity),
	}
}

func (rb *ringBuffer[T]) Cap() int {
	return len(rb.values)
}

func (rb *ringBuffer[T]) Len() int {
	return rb.size
}

func (rb *ringBuffer[T]) Append(value T) {
	if len(rb.values) == 0 {
		return
	}
	if rb.size < len(rb.values) {
		rb.values[(rb.start+rb.size)%len(rb.values)] = value
		rb.size++
		return
	}
	rb.values[rb.start] = value
	rb.start = (rb.start + 1) % len(rb.values)
}

func (rb *ringBuffer[T]) At(index int) T {
	return rb.values[(rb.start+index)%len(rb.values)]
}

func (rb *ringBuffer[T]) Snapshot() []T {
	items := make([]T, rb.size)
	for i := 0; i < rb.size; i++ {
		items[i] = rb.At(i)
	}
	return items
}

func (rb *ringBuffer[T]) Range(fn func(T) bool) {
	for i := 0; i < rb.size; i++ {
		if !fn(rb.At(i)) {
			return
		}
	}
}

type sentPacketSnapshot struct {
	seq       uint64
	timestamp uint64
}

type ChannelRx struct {
	ChannelID             uint64
	TotalPacketsReceived  uint64
	LastPacketSeqReceived uint64
}

func NewChannelRx(channelID uint64) *ChannelRx {
	return &ChannelRx{ChannelID: channelID}
}

func (cr *ChannelRx) AssignChannelID(channelID uint64) error {
	if channelID == 0 {
		return newError("invalid channel id")
	}
	if cr.ChannelID == 0 {
		cr.ChannelID = channelID
		return nil
	}
	if cr.ChannelID != channelID {
		return newError("channel id already assigned")
	}
	return nil
}

func (cr *ChannelRx) ProcessMessageReceived(msg ChannelDataMessage) error {
	cr.TotalPacketsReceived++
	if cr.TotalPacketsReceived == 1 || msg.ChannelSeq > cr.LastPacketSeqReceived {
		cr.LastPacketSeqReceived = msg.ChannelSeq
	}
	return nil
}

func (cr *ChannelRx) CreateControlMessage() (*ChannelControlMessage, error) {
	return &ChannelControlMessage{
		ChannelID:                  cr.ChannelID,
		TotalPacketReceived:        cr.TotalPacketsReceived,
		LastSequenceNumberReceived: cr.LastPacketSeqReceived,
	}, nil
}

type ChannelTx struct {
	ChannelID uint64
	NextSeq   uint64

	maxRewindableTimestampNum int
	// use a ring buffer to store packetSeq <-> time stamp look up

	maxRewindableControlMessageNum int
	// use a ring buffer to store the last few control message received

	sentPacketHistory ringBuffer[sentPacketSnapshot]
	controlHistory    ringBuffer[ChannelControlMessage]
}

func NewChannelTx(channelID uint64, maxRewindableTimestampNum int, maxRewindableControlMessageNum int) (*ChannelTx, error) {
	if maxRewindableTimestampNum < 0 {
		return nil, newError("invalid max rewindable timestamp number")
	}
	if maxRewindableControlMessageNum < 0 {
		return nil, newError("invalid max rewindable control message number")
	}

	tx := &ChannelTx{
		ChannelID:                      channelID,
		maxRewindableTimestampNum:      maxRewindableTimestampNum,
		maxRewindableControlMessageNum: maxRewindableControlMessageNum,
	}
	tx.sentPacketHistory = newRingBuffer[sentPacketSnapshot](maxRewindableTimestampNum)
	tx.controlHistory = newRingBuffer[ChannelControlMessage](maxRewindableControlMessageNum)
	return tx, nil
}

func (ct *ChannelTx) CreateDataMessage(data []byte, timestamp uint64) (*ChannelDataMessage, error) {
	payload := append([]byte(nil), data...)
	msg := &ChannelDataMessage{
		ChannelSeq: ct.NextSeq,
		Data:       payload,
	}
	ct.sentPacketHistory.Append(sentPacketSnapshot{
		seq:       ct.NextSeq,
		timestamp: timestamp,
	})
	ct.NextSeq++
	return msg, nil
}

func (ct *ChannelTx) AcceptControlMessage(ctrl ChannelControlMessage) error {
	if ctrl.ChannelID != ct.ChannelID {
		return newError("channel id mismatch")
	}
	ct.controlHistory.Append(ctrl)
	return nil
}

func (ct *ChannelTx) RemoteLastSeenMessageSenderTimestamp() (uint64, error) {
	bestControl, ok := ct.latestUsableControl()
	if !ok {
		return 0, newError("remote last seen timestamp unavailable")
	}
	timestamp, ok := ct.timestampForSeq(bestControl.LastSequenceNumberReceived)
	if !ok {
		return 0, newError("remote last seen timestamp unavailable")
	}
	return timestamp, nil
}

func (ct *ChannelTx) RemotePacketLossSinceTimestamp(since uint64) (sent, lost uint64, err error) {
	if ct.sentPacketHistory.Cap() == 0 || ct.sentPacketHistory.Len() == 0 {
		return 0, 0, newError("sent packet rewind history unavailable")
	}

	firstIdx, rewindable, ok := ct.firstPacketIndexSince(since)
	if !rewindable {
		return 0, 0, newError("timestamp outside rewind window")
	}
	if !ok {
		return 0, 0, nil
	}

	firstSeq := ct.sentPacketHistory.At(firstIdx).seq
	endControl, ok := ct.latestUsableControl()
	if !ok {
		return 0, 0, newError("remote control history unavailable")
	}
	if endControl.LastSequenceNumberReceived < firstSeq {
		return 0, 0, nil
	}

	for i := firstIdx; i < ct.sentPacketHistory.Len(); i++ {
		snapshot := ct.sentPacketHistory.At(i)
		if snapshot.seq > endControl.LastSequenceNumberReceived {
			break
		}
		sent++
	}

	startReceived := uint64(0)
	if startControl, ok := ct.bestControlBeforeSeq(firstSeq); ok {
		startReceived = startControl.TotalPacketReceived
	}
	if endControl.TotalPacketReceived < startReceived {
		return 0, 0, newError("remote control history regressed")
	}

	received := endControl.TotalPacketReceived - startReceived
	if received >= sent {
		return sent, 0, nil
	}
	return sent, sent - received, nil
}

func (ct *ChannelTx) latestUsableControl() (ChannelControlMessage, bool) {
	var (
		best  ChannelControlMessage
		found bool
	)
	ct.controlHistory.Range(func(ctrl ChannelControlMessage) bool {
		if ctrl.TotalPacketReceived == 0 {
			return true
		}
		if _, ok := ct.timestampForSeq(ctrl.LastSequenceNumberReceived); !ok {
			return true
		}
		if !found || ctrl.LastSequenceNumberReceived > best.LastSequenceNumberReceived {
			best = ctrl
			found = true
		}
		return true
	})
	return best, found
}

func (ct *ChannelTx) bestControlBeforeSeq(seq uint64) (ChannelControlMessage, bool) {
	var (
		best  ChannelControlMessage
		found bool
	)
	ct.controlHistory.Range(func(ctrl ChannelControlMessage) bool {
		if ctrl.TotalPacketReceived == 0 || ctrl.LastSequenceNumberReceived >= seq {
			return true
		}
		if !found || ctrl.LastSequenceNumberReceived > best.LastSequenceNumberReceived {
			best = ctrl
			found = true
		}
		return true
	})
	return best, found
}

func (ct *ChannelTx) timestampForSeq(seq uint64) (uint64, bool) {
	var (
		timestamp uint64
		found     bool
	)
	ct.sentPacketHistory.Range(func(snapshot sentPacketSnapshot) bool {
		if snapshot.seq == seq {
			timestamp = snapshot.timestamp
			found = true
			return false
		}
		return true
	})
	return timestamp, found
}

func (ct *ChannelTx) firstPacketIndexSince(since uint64) (idx int, rewindable bool, ok bool) {
	if ct.sentPacketHistory.Len() == 0 {
		return 0, false, false
	}
	oldest := ct.sentPacketHistory.At(0)
	if oldest.seq > 0 && since < oldest.timestamp {
		return 0, false, false
	}
	for i := 0; i < ct.sentPacketHistory.Len(); i++ {
		snapshot := ct.sentPacketHistory.At(i)
		if snapshot.timestamp >= since {
			return i, true, true
		}
	}
	return 0, true, false
}
