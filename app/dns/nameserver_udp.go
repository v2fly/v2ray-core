//go:build !confonly
// +build !confonly

package dns

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol/dns"
	udp_proto "github.com/ghxhy/v2ray-core/v5/common/protocol/udp"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/common/signal/pubsub"
	"github.com/ghxhy/v2ray-core/v5/common/task"
	dns_feature "github.com/ghxhy/v2ray-core/v5/features/dns"
	"github.com/ghxhy/v2ray-core/v5/features/routing"
	"github.com/ghxhy/v2ray-core/v5/transport/internet/udp"
)

// ClassicNameServer implemented traditional UDP DNS.
type ClassicNameServer struct {
	sync.RWMutex
	name      string
	address   net.Destination
	ips       map[string]record
	requests  map[uint16]dnsRequest
	pub       *pubsub.Service
	udpServer udp.DispatcherI
	cleanup   *task.Periodic
	reqID     uint32
}

// NewClassicNameServer creates udp server object for remote resolving.
func NewClassicNameServer(address net.Destination, dispatcher routing.Dispatcher) *ClassicNameServer {
	// default to 53 if unspecific
	if address.Port == 0 {
		address.Port = net.Port(53)
	}

	s := &ClassicNameServer{
		address:  address,
		ips:      make(map[string]record),
		requests: make(map[uint16]dnsRequest),
		pub:      pubsub.NewService(),
		name:     strings.ToUpper(address.String()),
	}
	s.cleanup = &task.Periodic{
		Interval: time.Minute,
		Execute:  s.Cleanup,
	}
	s.udpServer = udp.NewSplitDispatcher(dispatcher, s.HandleResponse)
	newError("DNS: created UDP client initialized for ", address.NetAddr()).AtInfo().WriteToLog()
	return s
}

// Name implements Server.
func (s *ClassicNameServer) Name() string {
	return s.name
}

// Cleanup clears expired items from cache
func (s *ClassicNameServer) Cleanup() error {
	now := time.Now()
	s.Lock()
	defer s.Unlock()

	if len(s.ips) == 0 && len(s.requests) == 0 {
		return newError(s.name, " nothing to do. stopping...")
	}

	for domain, record := range s.ips {
		if record.A != nil && record.A.Expire.Before(now) {
			record.A = nil
		}
		if record.AAAA != nil && record.AAAA.Expire.Before(now) {
			record.AAAA = nil
		}

		if record.A == nil && record.AAAA == nil {
			delete(s.ips, domain)
		} else {
			s.ips[domain] = record
		}
	}

	if len(s.ips) == 0 {
		s.ips = make(map[string]record)
	}

	for id, req := range s.requests {
		if req.expire.Before(now) {
			delete(s.requests, id)
		}
	}

	if len(s.requests) == 0 {
		s.requests = make(map[uint16]dnsRequest)
	}

	return nil
}

// HandleResponse handles udp response packet from remote DNS server.
func (s *ClassicNameServer) HandleResponse(ctx context.Context, packet *udp_proto.Packet) {
	ipRec, err := parseResponse(packet.Payload.Bytes())
	if err != nil {
		newError(s.name, " fail to parse responded DNS udp").AtError().WriteToLog()
		return
	}

	s.Lock()
	id := ipRec.ReqID
	req, ok := s.requests[id]
	if ok {
		// remove the pending request
		delete(s.requests, id)
	}
	s.Unlock()
	if !ok {
		newError(s.name, " cannot find the pending request").AtError().WriteToLog()
		return
	}

	var rec record
	switch req.reqType {
	case dnsmessage.TypeA:
		rec.A = ipRec
	case dnsmessage.TypeAAAA:
		rec.AAAA = ipRec
	}

	elapsed := time.Since(req.start)
	newError(s.name, " got answer: ", req.domain, " ", req.reqType, " -> ", ipRec.IP, " ", elapsed).AtInfo().WriteToLog()
	if len(req.domain) > 0 && (rec.A != nil || rec.AAAA != nil) {
		s.updateIP(req.domain, rec)
	}
}

func (s *ClassicNameServer) updateIP(domain string, newRec record) {
	s.Lock()

	newError(s.name, " updating IP records for domain:", domain).AtDebug().WriteToLog()
	rec := s.ips[domain]

	updated := false
	if isNewer(rec.A, newRec.A) {
		rec.A = newRec.A
		updated = true
	}
	if isNewer(rec.AAAA, newRec.AAAA) {
		rec.AAAA = newRec.AAAA
		updated = true
	}

	if updated {
		s.ips[domain] = rec
	}
	if newRec.A != nil {
		s.pub.Publish(domain+"4", nil)
	}
	if newRec.AAAA != nil {
		s.pub.Publish(domain+"6", nil)
	}
	s.Unlock()
	common.Must(s.cleanup.Start())
}

func (s *ClassicNameServer) newReqID() uint16 {
	return uint16(atomic.AddUint32(&s.reqID, 1))
}

func (s *ClassicNameServer) addPendingRequest(req *dnsRequest) {
	s.Lock()
	defer s.Unlock()

	id := req.msg.ID
	req.expire = time.Now().Add(time.Second * 8)
	s.requests[id] = *req
}

func (s *ClassicNameServer) sendQuery(ctx context.Context, domain string, clientIP net.IP, option dns_feature.IPOption) {
	newError(s.name, " querying DNS for: ", domain).AtDebug().WriteToLog(session.ExportIDToError(ctx))

	reqs := buildReqMsgs(domain, option, s.newReqID, genEDNS0Options(clientIP))

	for _, req := range reqs {
		s.addPendingRequest(req)
		b, _ := dns.PackMessage(req.msg)
		udpCtx := core.ToBackgroundDetachedContext(ctx)
		if inbound := session.InboundFromContext(ctx); inbound != nil {
			udpCtx = session.ContextWithInbound(udpCtx, inbound)
		}
		udpCtx = session.ContextWithContent(udpCtx, &session.Content{
			Protocol: "dns",
		})
		s.udpServer.Dispatch(udpCtx, s.address, b)
	}
}

func (s *ClassicNameServer) findIPsForDomain(domain string, option dns_feature.IPOption) ([]net.IP, error) {
	s.RLock()
	record, found := s.ips[domain]
	s.RUnlock()

	if !found {
		return nil, errRecordNotFound
	}

	var ips []net.Address
	var lastErr error
	if option.IPv4Enable {
		a, err := record.A.getIPs()
		if err != nil {
			lastErr = err
		}
		ips = append(ips, a...)
	}

	if option.IPv6Enable {
		aaaa, err := record.AAAA.getIPs()
		if err != nil {
			lastErr = err
		}
		ips = append(ips, aaaa...)
	}

	if len(ips) > 0 {
		return toNetIP(ips)
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, dns_feature.ErrEmptyResponse
}

// QueryIP implements Server.
func (s *ClassicNameServer) QueryIP(ctx context.Context, domain string, clientIP net.IP, option dns_feature.IPOption, disableCache bool) ([]net.IP, error) {
	fqdn := Fqdn(domain)

	if disableCache {
		newError("DNS cache is disabled. Querying IP for ", domain, " at ", s.name).AtDebug().WriteToLog()
	} else {
		ips, err := s.findIPsForDomain(fqdn, option)
		if err != errRecordNotFound {
			newError(s.name, " cache HIT ", domain, " -> ", ips).Base(err).AtDebug().WriteToLog()
			return ips, err
		}
	}

	// ipv4 and ipv6 belong to different subscription groups
	var sub4, sub6 *pubsub.Subscriber
	if option.IPv4Enable {
		sub4 = s.pub.Subscribe(fqdn + "4")
		defer sub4.Close()
	}
	if option.IPv6Enable {
		sub6 = s.pub.Subscribe(fqdn + "6")
		defer sub6.Close()
	}
	done := make(chan interface{})
	go func() {
		if sub4 != nil {
			select {
			case <-sub4.Wait():
			case <-ctx.Done():
			}
		}
		if sub6 != nil {
			select {
			case <-sub6.Wait():
			case <-ctx.Done():
			}
		}
		close(done)
	}()
	s.sendQuery(ctx, fqdn, clientIP, option)

	for {
		ips, err := s.findIPsForDomain(fqdn, option)
		if err != errRecordNotFound {
			return ips, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
	}
}
