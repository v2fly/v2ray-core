package http

import (
	"bytes"
	"errors"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

type version byte

type SniffHeader struct {
	host string
}

func (h *SniffHeader) Protocol() string {
	return "http1"
}

func (h *SniffHeader) Domain() string {
	return h.host
}

var (
	// refer to https://pkg.go.dev/net/http@master#pkg-constants
	methods = [...]string{"get", "post", "head", "put", "delete", "options", "connect", "patch", "trace"}

	errNotHTTPMethod = errors.New("not an HTTP method")
)

func beginWithHTTPMethod(b []byte) error {
	for _, m := range &methods {
		if len(b) >= len(m) && strings.EqualFold(string(b[:len(m)]), m) {
			return nil
		}

		if len(b) < len(m) {
			return common.ErrNoClue
		}
	}

	return errNotHTTPMethod
}

func SniffHTTP(b []byte) (*SniffHeader, error) {
	if err := beginWithHTTPMethod(b); err != nil {
		return nil, err
	}

	sh := &SniffHeader{}

	headers := bytes.Split(b, []byte{'\n'})
	for i := 1; i < len(headers); i++ {
		header := headers[i]
		if len(header) == 0 {
			break
		}
		parts := bytes.SplitN(header, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(string(parts[0]))
		if key == "host" {
			rawHost := strings.ToLower(string(bytes.TrimSpace(parts[1])))
			dest, err := ParseHost(rawHost, net.Port(80))
			if err != nil {
				return nil, err
			}
			sh.host = dest.Address.String()
		}
	}

	if len(sh.host) > 0 {
		return sh, nil
	}

	return nil, common.ErrNoClue
}
