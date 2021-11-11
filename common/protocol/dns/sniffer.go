package dns

import (
	"golang.org/x/net/dns/dnsmessage"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/errors"
)

var errNotDNS = errors.New("not dns")

type SniffHeader struct {
	domain string
}

func (s *SniffHeader) Protocol() string {
	return "dns"
}

func (s *SniffHeader) Domain() string {
	return s.domain
}

func SniffDNS(b []byte) (*SniffHeader, error) {
	var parser dnsmessage.Parser
	if common.Error2(parser.Start(b)) != nil {
		return nil, errNotDNS
	}
	question, err := parser.Question()
	if err != nil {
		return nil, errNotDNS
	}
	return &SniffHeader{domain: question.Name.String()}, nil
}
