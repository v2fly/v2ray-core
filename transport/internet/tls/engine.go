package tls

import (
	"errors"
	"fmt"
	"time"

	"github.com/miekg/dns"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
)

type Engine struct {
	config *Config
}

// ech key for ech enabled config. tls config use this for tls handshake.
var ECH string

func (e *Engine) fetchECHEvery(d time.Duration) {
	for range time.Tick(d) {
		ech, err := e.fetchECH()
		fmt.Printf("new ech = %s", ech)
		if err != nil {
			fmt.Println("failed to get ech")
			newError("failed to get ech").Base(err).AtError().WriteToLog()
		} else {
			ECH = ech
		}
	}
}

// only support cloudflare ech
func (e *Engine) fetchECH() (string, error) {
	c := dns.Client{Timeout: 10 * time.Second}

	d := dns.Fqdn("crypto.cloudflare.com")
	q := dns.Question{
		Name:   d,
		Qtype:  dns.TypeHTTPS,
		Qclass: dns.ClassINET,
	}

	dnsAddr := "1.1.1.1:53"
	if e.config.EchSetting != nil && e.config.EchSetting.DnsAddr != "" {
		dnsAddr = e.config.EchSetting.DnsAddr
	}

	r, _, err := c.Exchange(&dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{q},
	}, dnsAddr)
	if err != nil {
		return "", err
	}

	for _, v := range r.Answer {
		if vv, ok := v.(*dns.HTTPS); ok {
			for _, vvv := range vv.SVCB.Value {
				if vvv.Key().String() == "ech" {
					return vvv.String(), nil
				}
			}
		}
	}

	return "", errors.New("failed to found ech in response")
}

func (e *Engine) Client(conn net.Conn, opts ...security.Option) (security.Conn, error) {
	var options []Option
	for _, v := range opts {
		switch s := v.(type) {
		case security.OptionWithALPN:
			options = append(options, WithNextProto(s.ALPNs...))
		case security.OptionWithDestination:
			options = append(options, WithDestination(s.Dest))
		default:
			return nil, newError("unknown option")
		}
	}
	tlsConn := Client(conn, e.config.GetTLSConfig(options...))
	return tlsConn, nil
}

func NewTLSSecurityEngineFromConfig(config *Config) (security.Engine, error) {
	e := &Engine{config: config}

	// handle ech
	if config.EnableEch {
		ech, err := e.fetchECH()
		if err != nil {
			fmt.Println("failed to get first ech")
			newError("failed to get first ech").Base(err).AtError().WriteToLog()
		} else {
			ECH = ech
		}
		go e.fetchECHEvery(15 * time.Minute)
	}

	return e, nil
}
