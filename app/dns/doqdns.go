package dns

import (
	"context"
	"net/url"
	sync "sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/dns/dnsmessage"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	net "v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol/dns"
	"v2ray.com/core/common/session"
	"v2ray.com/core/common/signal/pubsub"
	"v2ray.com/core/common/task"
	dns_feature "v2ray.com/core/features/dns"
	"v2ray.com/core/transport/internet"
	"v2ray.com/core/transport/internet/quic"
	"v2ray.com/core/transport/internet/tls"
)

// DoQNameServer implemented DNS over QUIC
type DoQNameServer struct {
	sync.RWMutex
	ips         map[string]record
	pub         *pubsub.Service
	cleanup     *task.Periodic
	reqID       uint32
	clientIP    net.IP
	name        string
	destination net.Destination
}

// NewDoQNameServer creates DNS-over-QUIC client object for remote resolving
func NewDoQNameServer(url *url.URL, clientIP net.IP) (*DoQNameServer, error) {
	newError("DNS: created Remote DNS-over-QUIC client for ", url.String()).AtInfo().WriteToLog()
	if clientIP != nil {
		newError("DNS: Remote DNS-over-QUIC client ", url.String(), " uses clientip ", clientIP.String()).AtInfo().WriteToLog()
	}
	s, err := baseDOQNameServer(url, clientIP)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Name returns client name
func (s *DoQNameServer) Name() string {
	return s.name
}

// Cleanup clears expired items from cache
func (s *DoQNameServer) Cleanup() error {
	now := time.Now()
	s.Lock()
	defer s.Unlock()

	if len(s.ips) == 0 {
		return newError("nothing to do. stopping...")
	}

	for domain, record := range s.ips {
		if record.A != nil && record.A.Expire.Before(now) {
			record.A = nil
		}
		if record.AAAA != nil && record.AAAA.Expire.Before(now) {
			record.AAAA = nil
		}

		if record.A == nil && record.AAAA == nil {
			newError(s.name, " cleanup ", domain).AtDebug().WriteToLog()
			delete(s.ips, domain)
		} else {
			s.ips[domain] = record
		}
	}

	if len(s.ips) == 0 {
		s.ips = make(map[string]record)
	}

	return nil
}

func (s *DoQNameServer) updateIP(req *dnsRequest, ipRec *IPRecord) {
	elapsed := time.Since(req.start)

	s.Lock()
	rec := s.ips[req.domain]
	updated := false

	switch req.reqType {
	case dnsmessage.TypeA:
		if isNewer(rec.A, ipRec) {
			rec.A = ipRec
			updated = true
		}
	case dnsmessage.TypeAAAA:
		addr := make([]net.Address, 0)
		for _, ip := range ipRec.IP {
			if len(ip.IP()) == net.IPv6len {
				addr = append(addr, ip)
			}
		}
		ipRec.IP = addr
		if isNewer(rec.AAAA, ipRec) {
			rec.AAAA = ipRec
			updated = true
		}
	}
	newError(s.name, " got answer: ", req.domain, " ", req.reqType, " -> ", ipRec.IP, " ", elapsed).AtInfo().WriteToLog()

	if updated {
		s.ips[req.domain] = rec
	}
	switch req.reqType {
	case dnsmessage.TypeA:
		s.pub.Publish(req.domain+"4", nil)
	case dnsmessage.TypeAAAA:
		s.pub.Publish(req.domain+"6", nil)
	}
	s.Unlock()
	common.Must(s.cleanup.Start())
}

func (s *DoQNameServer) newReqID() uint16 {
	return uint16(atomic.AddUint32(&s.reqID, 1))
}

func (s *DoQNameServer) sendQuery(ctx context.Context, domain string, option IPOption) {
	newError(s.name, " querying: ", domain).AtInfo().WriteToLog(session.ExportIDToError(ctx))

	reqs := buildReqMsgs(domain, option, s.newReqID, genEDNS0Options(s.clientIP))

	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(time.Second * 5)
	}

	for _, req := range reqs {
		go func(r *dnsRequest) {
			// generate new context for each req, using same context
			// may cause reqs all aborted if any one encounter an error
			dnsCtx := context.Background()

			// reserve internal dns server requested Inbound
			if inbound := session.InboundFromContext(ctx); inbound != nil {
				dnsCtx = session.ContextWithInbound(dnsCtx, inbound)
			}

			dnsCtx = session.ContextWithContent(dnsCtx, &session.Content{
				Protocol:      "quic",
				SkipRoutePick: true,
			})

			var cancel context.CancelFunc
			dnsCtx, cancel = context.WithDeadline(dnsCtx, deadline)
			defer cancel()

			b, err := dns.PackMessage(r.msg)
			if err != nil {
				newError("failed to pack dns query").Base(err).AtError().WriteToLog()
				return
			}

			conn, err := quic.Dial(dnsCtx, s.destination, &internet.MemoryStreamConfig{
				ProtocolName: "quic",
				SecurityType: "tls",
				SecuritySettings: &tls.Config{
					AllowInsecure: true,
				},
			})
			if err != nil {
				newError("failed to open quic session").Base(err).AtError().WriteToLog()
				return
			}

			_, err = conn.Write(b.Bytes())
			if err != nil {
				newError("failed to send query").Base(err).AtError().WriteToLog()
				return
			}

			conn.Close()

			respBuf, err := buf.ReadBuffer(conn)
			if err != nil {
				newError("failed to read response").Base(err).AtError().WriteToLog()
				return
			}

			rec, err := parseResponse(respBuf.Bytes())
			if err != nil {
				newError("failed to handle response").Base(err).AtError().WriteToLog()
				return
			}
			s.updateIP(r, rec)
		}(req)
	}
}

func (s *DoQNameServer) findIPsForDomain(domain string, option IPOption) ([]net.IP, error) {
	s.RLock()
	record, found := s.ips[domain]
	s.RUnlock()

	if !found {
		return nil, errRecordNotFound
	}

	var ips []net.Address
	var lastErr error
	if option.IPv6Enable && record.AAAA != nil && record.AAAA.RCode == dnsmessage.RCodeSuccess {
		aaaa, err := record.AAAA.getIPs()
		if err != nil {
			lastErr = err
		}
		ips = append(ips, aaaa...)
	}

	if option.IPv4Enable && record.A != nil && record.A.RCode == dnsmessage.RCodeSuccess {
		a, err := record.A.getIPs()
		if err != nil {
			lastErr = err
		}
		ips = append(ips, a...)
	}

	if len(ips) > 0 {
		return toNetIP(ips), nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	if (option.IPv4Enable && record.A != nil) || (option.IPv6Enable && record.AAAA != nil) {
		return nil, dns_feature.ErrEmptyResponse
	}

	return nil, errRecordNotFound
}

// QueryIP is called from dns.Server->queryIPTimeout
func (s *DoQNameServer) QueryIP(ctx context.Context, domain string, option IPOption) ([]net.IP, error) {
	fqdn := Fqdn(domain)

	ips, err := s.findIPsForDomain(fqdn, option)
	if err != errRecordNotFound {
		newError(s.name, " cache HIT ", domain, " -> ", ips).Base(err).AtDebug().WriteToLog()
		return ips, err
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
	s.sendQuery(ctx, fqdn, option)

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

func baseDOQNameServer(url *url.URL, clientIP net.IP) (*DoQNameServer, error) {
	var err error
	port := net.Port(784)
	if url.Port() != "" {
		port, err = net.PortFromString(url.Port())
		if err != nil {
			return nil, err
		}
	}
	dest := net.UDPDestination(net.DomainAddress(url.Hostname()), port)

	s := &DoQNameServer{
		ips:         make(map[string]record),
		clientIP:    clientIP,
		pub:         pubsub.NewService(),
		name:        url.String(),
		destination: dest,
	}
	s.cleanup = &task.Periodic{
		Interval: time.Minute,
		Execute:  s.Cleanup,
	}

	return s, nil
}
