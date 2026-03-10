package stun

// Mostly Machine Generated Code
import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/pion/stun/v3"

	"github.com/v2fly/v2ray-core/v5/common/task"
)

type NATDependantType int

const (
	Unknown NATDependantType = iota
	Independent
	EndpointDependent
	EndpointPortDependent
	EndpointPortDependentPinned
)

type NATYesOrNoUnknownType int

const (
	NATYesOrNoUnknownType_Unknown NATYesOrNoUnknownType = iota
	NATYesOrNoUnknownType_Yes
	NATYesOrNoUnknownType_No
)

type NATTypeTest struct {
	newStunConn     func() (net.PacketConn, error)
	testsTranscript []TestConducted
	transcriptMux   sync.Mutex

	Timeout  time.Duration
	Attempts int

	FilterBehaviour  NATDependantType
	MappingBehaviour NATDependantType
	HairpinBehaviour NATYesOrNoUnknownType

	// Calculated values from testsTranscript
	PreserveSourcePortWhenSourceNATMapping NATYesOrNoUnknownType
	SingleSourceIPSourceNATMapping         NATYesOrNoUnknownType
	// PreserveSourceIPPortWhenDestNATMapping
	// means when receiving packets,
	// whether the real source address is preserved in the reply message
	// some time a bad proxy would fill in a default value rather the real remote address
	// this can be detected when asking remote server to reply from a different ip or port
	PreserveSourceIPPortWhenDestNATMapping NATYesOrNoUnknownType

	TestServer net.Addr

	SourceIPs []net.IP
}

func NewNATTypeTest(newStunConn func() (net.PacketConn, error), testServer net.Addr, timeout time.Duration, attempts int) *NATTypeTest {
	return &NATTypeTest{
		newStunConn: newStunConn,
		Timeout:     timeout,
		Attempts:    attempts,
		TestServer:  testServer,
	}
}

type TestConducted struct {
	Req         stun.Message
	ReqSentTo   net.Addr
	ReqSentFrom net.Addr
	Resp        *stun.Message
	RespFrom    net.Addr
}

func changeRequestSetter(changeIP, changePort bool) stun.RawAttribute {
	val := make([]byte, 4)
	var flags uint32
	if changeIP {
		flags |= 0x04
	}
	if changePort {
		flags |= 0x02
	}
	binary.BigEndian.PutUint32(val, flags)
	return stun.RawAttribute{
		Type:  stun.AttrChangeRequest,
		Value: val,
	}
}

func startBackgroundReader(conn *StunClientConn) {
	go func() {
		buf := make([]byte, 1500)
		for {
			_, _, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
		}
	}()
}

func (t *NATTypeTest) recordTransaction(tc TestConducted) {
	t.transcriptMux.Lock()
	defer t.transcriptMux.Unlock()
	t.testsTranscript = append(t.testsTranscript, tc)
}

// doTransactionWithRetry sends multiple STUN requests at once (each with a fresh
// transaction ID) and waits for the first response within a single timeout window.
// This avoids sequential retry delays caused by UDP packet loss.
// Non-timeout errors from sending are returned immediately.
func (t *NATTypeTest) doTransactionWithRetry(conn *StunClientConn, localAddr net.Addr, dest net.Addr, attempts int, setters ...stun.Setter) (*stun.Message, net.Addr, error) {
	type result struct {
		msg  stun.Message
		addr net.Addr
	}
	ch := make(chan result, attempts)

	var txIDs []stunTransactionID
	var firstMsg *stun.Message

	for i := 0; i < attempts; i++ {
		msg := stun.MustBuild(setters...)
		if i == 0 {
			firstMsg = msg
		}

		_, _, err := conn.ExecuteSTUNMessageAsync(*msg, dest, func(_ [stun.TransactionIDSize]byte, respMsg stun.Message, respAddr net.Addr) {
			ch <- result{msg: respMsg, addr: respAddr}
		})
		if err != nil {
			for _, id := range txIDs {
				conn.processor.CancelTransaction(id)
			}
			return nil, nil, err
		}
		txIDs = append(txIDs, msg.TransactionID)
	}

	// Wait for first response or timeout
	var resp *result
	select {
	case r := <-ch:
		resp = &r
	case <-time.After(t.Timeout):
	}

	// Cancel all remaining pending transactions
	for _, id := range txIDs {
		conn.processor.CancelTransaction(id)
	}

	// Record result
	if resp != nil {
		respMsg := resp.msg
		t.recordTransaction(TestConducted{
			Req:         *firstMsg,
			ReqSentTo:   dest,
			ReqSentFrom: localAddr,
			Resp:        &respMsg,
			RespFrom:    resp.addr,
		})
		return &respMsg, resp.addr, nil
	}

	t.recordTransaction(TestConducted{
		Req:         *firstMsg,
		ReqSentTo:   dest,
		ReqSentFrom: localAddr,
	})
	return nil, nil, ErrTimeout
}

// TestFilterBehaviour determines NAT filtering behavior per RFC 5780 Section 4.4.
func (t *NATTypeTest) TestFilterBehaviour() error {
	rawConn, err := t.newStunConn()
	if err != nil {
		return err
	}

	conn, err := NewStunClientConn(rawConn)
	if err != nil {
		rawConn.Close()
		return err
	}
	defer conn.Close()
	localAddr := rawConn.LocalAddr()
	startBackgroundReader(conn)

	// Test I: Regular binding to confirm connectivity and get OTHER-ADDRESS
	resp1, _, err := t.doTransactionWithRetry(conn, localAddr, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}

	// Check if server supports RFC 5780 (OTHER-ADDRESS).
	// Without it, CHANGE-REQUEST results are unreliable.
	var filterOtherAddr stun.OtherAddress
	if err := filterOtherAddr.GetFrom(resp1); err != nil {
		t.FilterBehaviour = Unknown
		return nil
	}

	// Test II: Request server to respond from different IP and port
	_, _, err = t.doTransactionWithRetry(conn, localAddr, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest, changeRequestSetter(true, true))
	if err == nil {
		t.FilterBehaviour = Independent
		return nil
	}
	if !errors.Is(err, ErrTimeout) {
		return err
	}

	// Test III: Request server to respond from different port only
	_, _, err = t.doTransactionWithRetry(conn, localAddr, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest, changeRequestSetter(false, true))
	if err == nil {
		t.FilterBehaviour = EndpointDependent
		return nil
	}
	if !errors.Is(err, ErrTimeout) {
		return err
	}

	// Test IV: Check if sending outbound UDP can open the filter for a new endpoint.
	// Send a binding to the alternative address to create a NAT filter entry,
	// then ask the original server to reply from that alternative address.
	// If the response arrives, the filter can be opened by outbound packets.
	altAddr := &net.UDPAddr{IP: filterOtherAddr.IP, Port: filterOtherAddr.Port}

	// Send binding to alt address to open the NAT filter for that endpoint
	_, _, err = t.doTransactionWithRetry(conn, localAddr, altAddr, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil && !errors.Is(err, ErrTimeout) {
		return err
	}

	// Now ask original server to reply from the alternative address
	_, _, err = t.doTransactionWithRetry(conn, localAddr, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest, changeRequestSetter(true, true))
	if err == nil {
		t.FilterBehaviour = EndpointPortDependent
		return nil
	}
	if !errors.Is(err, ErrTimeout) {
		return err
	}

	t.FilterBehaviour = EndpointPortDependentPinned
	return nil
}

// TestMappingBehaviour determines NAT mapping behavior per RFC 5780 Section 4.3.
func (t *NATTypeTest) TestMappingBehaviour() error {
	rawConn, err := t.newStunConn()
	if err != nil {
		return err
	}

	conn, err := NewStunClientConn(rawConn)
	if err != nil {
		rawConn.Close()
		return err
	}
	defer conn.Close()
	localAddr := rawConn.LocalAddr()
	startBackgroundReader(conn)

	// Test I: Regular binding to primary server
	resp1, _, err := t.doTransactionWithRetry(conn, localAddr, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}

	var mappedAddr1 stun.XORMappedAddress
	if err := mappedAddr1.GetFrom(resp1); err != nil {
		return err
	}

	var otherAddr stun.OtherAddress
	if err := otherAddr.GetFrom(resp1); err != nil {
		// Server does not support RFC 5780 (no OTHER-ADDRESS), cannot test mapping
		t.MappingBehaviour = Unknown
		return nil
	}

	// Test II: From same socket, binding to OTHER-ADDRESS (different IP and port)
	altAddr := &net.UDPAddr{IP: otherAddr.IP, Port: otherAddr.Port}
	resp2, _, err := t.doTransactionWithRetry(conn, localAddr, altAddr, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}

	var mappedAddr2 stun.XORMappedAddress
	if err := mappedAddr2.GetFrom(resp2); err != nil {
		return err
	}

	if mappedAddr1.String() == mappedAddr2.String() {
		t.MappingBehaviour = Independent
		return nil
	}

	// Test III: From same socket, binding to (other IP, original port)
	serverUDP, ok := t.TestServer.(*net.UDPAddr)
	if !ok {
		return errors.New("TestServer is not a UDP address")
	}
	altAddr2 := &net.UDPAddr{IP: otherAddr.IP, Port: serverUDP.Port}
	resp3, _, err := t.doTransactionWithRetry(conn, localAddr, altAddr2, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}

	var mappedAddr3 stun.XORMappedAddress
	if err := mappedAddr3.GetFrom(resp3); err != nil {
		return err
	}

	if mappedAddr2.String() == mappedAddr3.String() {
		t.MappingBehaviour = EndpointDependent
	} else {
		t.MappingBehaviour = EndpointPortDependent
	}
	return nil
}

// TestHairpinBehaviour determines if the NAT supports hairpinning per RFC 5780 Section 4.5.
// Both sockets must first get their mapped addresses via STUN, then send to each other's
// mapped address. This ensures the NAT filter is opened for the peer's mapped address
// before the hairpin test packet arrives, avoiding false negatives from filtering.
func (t *NATTypeTest) TestHairpinBehaviour() error {
	// Socket 1: get mapped address
	rawConn1, err := t.newStunConn()
	if err != nil {
		return err
	}
	conn1, err := NewStunClientConn(rawConn1)
	if err != nil {
		rawConn1.Close()
		return err
	}
	defer conn1.Close()
	localAddr1 := rawConn1.LocalAddr()
	startBackgroundReader(conn1)

	resp1, _, err := t.doTransactionWithRetry(conn1, localAddr1, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}
	var mappedAddr1 stun.XORMappedAddress
	if err := mappedAddr1.GetFrom(resp1); err != nil {
		return err
	}
	selfAddr1 := &net.UDPAddr{IP: mappedAddr1.IP, Port: mappedAddr1.Port}

	// Socket 2: get mapped address
	rawConn2, err := t.newStunConn()
	if err != nil {
		return err
	}
	conn2, err := NewStunClientConn(rawConn2)
	if err != nil {
		rawConn2.Close()
		return err
	}
	defer conn2.Close()
	localAddr2 := rawConn2.LocalAddr()
	startBackgroundReader(conn2)

	resp2, _, err := t.doTransactionWithRetry(conn2, localAddr2, t.TestServer, t.Attempts,
		stun.TransactionID, stun.BindingRequest)
	if err != nil {
		return err
	}
	var mappedAddr2 stun.XORMappedAddress
	if err := mappedAddr2.GetFrom(resp2); err != nil {
		return err
	}
	selfAddr2 := &net.UDPAddr{IP: mappedAddr2.IP, Port: mappedAddr2.Port}

	// Socket 1 sends to MA2 to open the NAT filter for MA2 on socket 1's side.
	// Without this, a hairpinned packet from socket 2 (appearing as MA2) would be
	// filtered by endpoint-dependent filtering on socket 1.
	openMsg := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	openMsg.Encode()
	conn1.PacketConn.WriteTo(openMsg.Raw, selfAddr2)

	// Build hairpin test messages: register on conn1's processor, send from conn2.
	// Hairpinned packets arrive at conn1 from MA2 (now allowed by filter).
	type result struct {
		msg  stun.Message
		addr net.Addr
	}
	ch := make(chan result, t.Attempts)

	var txIDs []stunTransactionID
	var firstMsg *stun.Message
	for i := 0; i < t.Attempts; i++ {
		msg := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
		if i == 0 {
			firstMsg = msg
		}
		msg.Encode()

		// Register on conn1's processor (hairpinned packet arrives at conn1)
		conn1.processor.AddPendingTransactionListener(msg.TransactionID, func(_ [stun.TransactionIDSize]byte, respMsg stun.Message, respAddr net.Addr) {
			ch <- result{msg: respMsg, addr: respAddr}
		})
		txIDs = append(txIDs, msg.TransactionID)

		// Send from conn2 to MA1 (socket 1's mapped address)
		if _, err := conn2.PacketConn.WriteTo(msg.Raw, selfAddr1); err != nil {
			for _, id := range txIDs {
				conn1.processor.CancelTransaction(id)
			}
			return err
		}
	}

	// Wait for hairpinned packet on conn1
	var respResult *result
	select {
	case r := <-ch:
		respResult = &r
	case <-time.After(t.Timeout):
	}

	// Cancel all remaining pending transactions
	for _, id := range txIDs {
		conn1.processor.CancelTransaction(id)
	}

	t.HairpinBehaviour = NATYesOrNoUnknownType_No

	if respResult != nil {
		respMsg := respResult.msg
		t.recordTransaction(TestConducted{
			Req:         *firstMsg,
			ReqSentTo:   selfAddr1,
			ReqSentFrom: localAddr2,
			Resp:        &respMsg,
			RespFrom:    respResult.addr,
		})
		if respMsg.Type == stun.BindingRequest {
			t.HairpinBehaviour = NATYesOrNoUnknownType_Yes
		}
		return nil
	}

	t.recordTransaction(TestConducted{
		Req:         *firstMsg,
		ReqSentTo:   selfAddr1,
		ReqSentFrom: localAddr2,
	})
	return nil
}

// TestAll runs all NAT behavior tests in parallel, then calculates derived values.
func (t *NATTypeTest) TestAll() error {
	err := task.Run(context.Background(),
		t.TestFilterBehaviour,
		t.TestMappingBehaviour,
		t.TestHairpinBehaviour,
	)
	if err != nil {
		return err
	}
	return t.CalcReminderValues()
}

// CalcReminderValues derives additional NAT properties from the collected test transcripts.
func (t *NATTypeTest) CalcReminderValues() error {
	t.transcriptMux.Lock()
	transcripts := make([]TestConducted, len(t.testsTranscript))
	copy(transcripts, t.testsTranscript)
	t.transcriptMux.Unlock()

	type addrKey struct {
		ip   string
		port int
	}

	var mappedAddrs []addrKey
	for _, tc := range transcripts {
		if tc.Resp == nil {
			continue
		}
		var addr stun.XORMappedAddress
		if err := addr.GetFrom(tc.Resp); err != nil {
			continue
		}
		mappedAddrs = append(mappedAddrs, addrKey{ip: addr.IP.String(), port: addr.Port})
	}

	// Collect unique mapped source IPs
	seenIPs := make(map[string]struct{})
	t.SourceIPs = nil
	for _, m := range mappedAddrs {
		if _, ok := seenIPs[m.ip]; !ok {
			seenIPs[m.ip] = struct{}{}
			t.SourceIPs = append(t.SourceIPs, net.ParseIP(m.ip))
		}
	}

	// Need at least 2 mapped addresses to draw any meaningful comparison
	if len(mappedAddrs) < 2 {
		t.SingleSourceIPSourceNATMapping = NATYesOrNoUnknownType_Unknown
		t.PreserveSourceIPPortWhenDestNATMapping = NATYesOrNoUnknownType_Unknown
		t.PreserveSourcePortWhenSourceNATMapping = NATYesOrNoUnknownType_Unknown
		return nil
	}

	// SingleSourceIPSourceNATMapping: check if all mapped IPs are the same
	allSameIP := true
	for _, m := range mappedAddrs[1:] {
		if m.ip != mappedAddrs[0].ip {
			allSameIP = false
			break
		}
	}
	if allSameIP {
		t.SingleSourceIPSourceNATMapping = NATYesOrNoUnknownType_Yes
	} else {
		t.SingleSourceIPSourceNATMapping = NATYesOrNoUnknownType_No
	}

	allSendToMatchRespFrom := true
	validPairCount := 0
	for _, tc := range transcripts {
		if tc.Resp == nil || tc.RespFrom == nil || tc.ReqSentTo == nil {
			continue
		}
		if value, ok := tc.Req.Attributes.Get(stun.AttrChangeRequest); ok {
			if len(value.Value) != 4 || (value.Value[0] == 0 && value.Value[1] == 0 && value.Value[2] == 0 && value.Value[3] == 0) {
				continue
			}
		} else {
			continue
		}
		validPairCount++
		if tc.RespFrom.String() != tc.ReqSentTo.String() {
			allSendToMatchRespFrom = false
			break
		}
	}
	if validPairCount < 1 {
		t.PreserveSourceIPPortWhenDestNATMapping = NATYesOrNoUnknownType_Unknown
	} else if allSendToMatchRespFrom {
		t.PreserveSourceIPPortWhenDestNATMapping = NATYesOrNoUnknownType_No
	} else {
		t.PreserveSourceIPPortWhenDestNATMapping = NATYesOrNoUnknownType_Yes
	}

	// PreserveSourcePortWhenSourceNATMapping: check if mapped port matches local source port
	preserves := true
	validCount := 0
	for _, tc := range transcripts {
		if tc.Resp == nil || tc.ReqSentFrom == nil {
			continue
		}
		localUDP, ok := tc.ReqSentFrom.(*net.UDPAddr)
		if !ok || localUDP.Port == 0 {
			continue
		}
		var addr stun.XORMappedAddress
		if err := addr.GetFrom(tc.Resp); err != nil {
			continue
		}
		validCount++
		if addr.Port != localUDP.Port {
			preserves = false
		}
	}
	if validCount >= 2 {
		if preserves {
			t.PreserveSourcePortWhenSourceNATMapping = NATYesOrNoUnknownType_Yes
		} else {
			t.PreserveSourcePortWhenSourceNATMapping = NATYesOrNoUnknownType_No
		}
	} else {
		t.PreserveSourcePortWhenSourceNATMapping = NATYesOrNoUnknownType_Unknown
	}

	return nil
}
