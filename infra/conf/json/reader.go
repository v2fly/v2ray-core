package json

import (
	"io"

	"github.com/v2fly/v2ray-core/v5/common/buf"
)

// State is the internal state of parser.
type State byte

const (
	StateContent State = iota
	StateEscape
	StateDoubleQuote
	StateDoubleQuoteEscape
	StateSingleQuote
	StateSingleQuoteEscape
	StateComment
	StateSlash
	StateMultilineComment
	StateMultilineCommentStar
)

// Reader is a reader for filtering comments.
// It supports Java style single and multi line comment syntax, and Python style single line comment syntax.
type Reader struct {
	io.Reader

	state   State
	pending []byte
	br      *buf.BufferedReader
}

// Read implements io.Reader.Read().
func (v *Reader) Read(b []byte) (int, error) {
	if v.br == nil {
		v.br = &buf.BufferedReader{Reader: buf.NewReader(v.Reader)}
	}

	p := b[:0]
	for len(p) < len(b) {
		if len(v.pending) > 0 {
			max := len(b) - len(p)
			if max > len(v.pending) {
				max = len(v.pending)
			}
			p = append(p, v.pending[:max]...)
			v.pending = v.pending[max:]
			continue
		}

		x, err := v.br.ReadByte()
		if err != nil {
			if len(p) == 0 {
				return 0, err
			}
			return len(p), nil
		}
		switch v.state {
		case StateContent:
			switch x {
			case '"':
				v.state = StateDoubleQuote
				p = append(p, x)
			case '\'':
				v.state = StateSingleQuote
				p = append(p, x)
			case '\\':
				v.state = StateEscape
				p = append(p, x)
			case '#':
				v.state = StateComment
			case '/':
				v.state = StateSlash
			default:
				p = append(p, x)
			}
		case StateEscape:
			p = append(p, x)
			v.state = StateContent
		case StateDoubleQuote:
			switch x {
			case '"':
				v.state = StateContent
				p = append(p, x)
			case '\\':
				v.state = StateDoubleQuoteEscape
				p = append(p, x)
			default:
				p = append(p, x)
			}
		case StateDoubleQuoteEscape:
			p = append(p, x)
			v.state = StateDoubleQuote
		case StateSingleQuote:
			switch x {
			case '\'':
				v.state = StateContent
				p = append(p, x)
			case '\\':
				v.state = StateSingleQuoteEscape
				p = append(p, x)
			default:
				p = append(p, x)
			}
		case StateSingleQuoteEscape:
			p = append(p, x)
			v.state = StateSingleQuote
		case StateComment:
			if x == '\n' {
				v.state = StateContent
				p = append(p, x)
			}
		case StateSlash:
			switch x {
			case '/':
				v.state = StateComment
			case '*':
				v.state = StateMultilineComment
			default:
				v.state = StateContent
				v.pending = append(v.pending, x)
				p = append(p, '/')
			}
		case StateMultilineComment:
			switch x {
			case '*':
				v.state = StateMultilineCommentStar
			case '\n':
				p = append(p, x)
			}
		case StateMultilineCommentStar:
			switch x {
			case '/':
				v.state = StateContent
			case '*':
				// Stay
			case '\n':
				p = append(p, x)
			default:
				v.state = StateMultilineComment
			}
		default:
			panic("Unknown state.")
		}
	}
	return len(p), nil
}
