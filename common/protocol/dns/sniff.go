package dns

import (
	"golang.org/x/net/dns/dnsmessage"

	"github.com/v2fly/v2ray-core/v4/common"
)

type SniffHeader struct {
	domain string
}

func (h *SniffHeader) Protocol() string {
	return "dns"
}

func (h *SniffHeader) Domain() string {
	return h.domain
}

func ParseIPQuery(b []byte) (r bool, domain string, id uint16, qType dnsmessage.Type) {
	var parser dnsmessage.Parser
	header, err := parser.Start(b)
	if err != nil {
		newError("parser start").Base(err).WriteToLog()
		return
	}

	id = header.ID
	q, err := parser.Question()
	if err != nil {
		newError("question").Base(err).WriteToLog()
		return
	}
	qType = q.Type
	if qType != dnsmessage.TypeA && qType != dnsmessage.TypeAAAA {
		return
	}

	domain = q.Name.String()
	r = true
	return
}

func SniffDNS(b []byte) (*SniffHeader, error) {
	h := &SniffHeader{}

	isIPQuery, domain, _, _ := ParseIPQuery(b)

	if isIPQuery {
		h.domain = domain
		return h, nil
	}

	return nil, common.ErrNoClue
}
