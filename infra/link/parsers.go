package link

import (
	"fmt"
	"net/url"
	"strings"

	"v2ray.com/core/common/errors"
)

// ParseFunc is parser function to load v2ray links, like "vmess://..."
type ParseFunc func(input string) (Link, error)

// Parser is parser load v2ray links with specified schemes
type Parser struct {
	Name   string
	Scheme []string
	Parse  ParseFunc
}

var (
	parsers = make(map[string][]*Parser)
)

// RegisterParser add a new ConfigLoader.
func RegisterParser(parser *Parser) error {
	for _, scheme := range parser.Scheme {
		s := strings.ToLower(scheme)
		ps, _ := parsers[s]
		if len(ps) == 0 {
			ps = make([]*Parser, 0)
		}
		parsers[s] = append(ps, parser)
	}

	return nil
}

func getParsers(link string) ([]*Parser, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, errors.New("invalid link")
	}
	s := strings.ToLower(u.Scheme)
	ps, _ := parsers[s]
	if len(ps) == 0 {
		return nil, fmt.Errorf("unsupported link scheme: %s", u.Scheme)
	}
	return ps, nil
}
