package gdocsviewer

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"strings"
)

const maxGeneratedOriginURLReplacementBytes = 256

type originURLPatternAtom struct {
	choices []byte
	count   int
}

func renderOriginURL(config *ClientConfig) (string, error) {
	if config == nil || config.OriginUrl == "" {
		return "", newError("origin_url is required")
	}
	originURL := config.OriginUrl
	for _, rule := range config.OriginUrlReplacementRules {
		value, err := generateOriginURLReplacement(rule, crand.Reader)
		if err != nil {
			return "", err
		}
		originURL = strings.ReplaceAll(originURL, "{"+rule.Name+"}", value)
	}
	return strings.TrimRight(originURL, "/"), nil
}

func generateOriginURLReplacement(rule *OriginUrlReplacementRule, random io.Reader) (string, error) {
	if rule == nil {
		return "", newError("origin_url_replacement_rules contains nil rule")
	}
	if err := validateOriginURLReplacementName(rule.Name); err != nil {
		return "", err
	}
	compiled, length, err := compileOriginURLPattern(rule.Pattern)
	if err != nil {
		return "", err
	}
	out := make([]byte, 0, length)
	for _, atom := range compiled {
		for i := 0; i < atom.count; i++ {
			index, err := randomIndex(random, len(atom.choices))
			if err != nil {
				return "", newError("unable to generate origin URL replacement").Base(err)
			}
			out = append(out, atom.choices[index])
		}
	}
	return string(out), nil
}

func validateOriginURLReplacementName(name string) error {
	if name == "" {
		return newError("origin URL replacement rule name is required")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' {
			continue
		}
		return newError("invalid origin URL replacement rule name: ", name)
	}
	return nil
}

func compileOriginURLPattern(pattern string) ([]originURLPatternAtom, int, error) {
	if pattern == "" {
		return nil, 0, newError("origin URL replacement pattern is required")
	}
	var atoms []originURLPatternAtom
	totalLength := 0
	for i := 0; i < len(pattern); {
		var atom originURLPatternAtom
		var err error
		switch pattern[i] {
		case '[':
			atom.choices, i, err = parseOriginURLPatternClass(pattern, i)
			if err != nil {
				return nil, 0, err
			}
		case '\\':
			if i+1 >= len(pattern) {
				return nil, 0, newError("dangling escape in origin URL replacement pattern")
			}
			atom.choices = []byte{pattern[i+1]}
			i += 2
		case '(', ')', '|', '*', '+', '?':
			return nil, 0, newError("unsupported origin URL replacement pattern operator: ", string(pattern[i]))
		case '{', '}':
			return nil, 0, newError("unexpected repeat marker in origin URL replacement pattern")
		default:
			atom.choices = []byte{pattern[i]}
			i++
		}
		atom.count = 1
		if i < len(pattern) && pattern[i] == '{' {
			atom.count, i, err = parseOriginURLPatternRepeat(pattern, i)
			if err != nil {
				return nil, 0, err
			}
		}
		totalLength += atom.count
		if totalLength > maxGeneratedOriginURLReplacementBytes {
			return nil, 0, newError("origin URL replacement output exceeds limit: ", maxGeneratedOriginURLReplacementBytes)
		}
		atoms = append(atoms, atom)
	}
	return atoms, totalLength, nil
}

func parseOriginURLPatternClass(pattern string, pos int) ([]byte, int, error) {
	var choices []byte
	i := pos + 1
	for i < len(pattern) {
		if pattern[i] == ']' {
			if len(choices) == 0 {
				return nil, 0, newError("empty character class in origin URL replacement pattern")
			}
			return choices, i + 1, nil
		}
		start, next, escaped, err := readOriginURLClassChar(pattern, i)
		if err != nil {
			return nil, 0, err
		}
		if !escaped && next < len(pattern) && pattern[next] == '-' && next+1 < len(pattern) && pattern[next+1] != ']' {
			end, afterEnd, _, err := readOriginURLClassChar(pattern, next+1)
			if err != nil {
				return nil, 0, err
			}
			if end < start {
				return nil, 0, newError("invalid descending range in origin URL replacement pattern")
			}
			for c := int(start); c <= int(end); c++ {
				choices = append(choices, byte(c))
			}
			i = afterEnd
			continue
		}
		choices = append(choices, start)
		i = next
	}
	return nil, 0, newError("unterminated character class in origin URL replacement pattern")
}

func readOriginURLClassChar(pattern string, pos int) (byte, int, bool, error) {
	if pattern[pos] == '\\' {
		if pos+1 >= len(pattern) {
			return 0, 0, false, newError("dangling escape in origin URL replacement character class")
		}
		return pattern[pos+1], pos + 2, true, nil
	}
	return pattern[pos], pos + 1, false, nil
}

func parseOriginURLPatternRepeat(pattern string, pos int) (int, int, error) {
	i := pos + 1
	if i >= len(pattern) || pattern[i] < '0' || pattern[i] > '9' {
		return 0, 0, newError("origin URL replacement repeat must be a fixed count")
	}
	count := 0
	for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
		count = count*10 + int(pattern[i]-'0')
		if count > maxGeneratedOriginURLReplacementBytes {
			return 0, 0, newError("origin URL replacement repeat exceeds limit: ", maxGeneratedOriginURLReplacementBytes)
		}
		i++
	}
	if i >= len(pattern) || pattern[i] != '}' {
		return 0, 0, newError("unterminated repeat in origin URL replacement pattern")
	}
	if count == 0 {
		return 0, 0, newError("origin URL replacement repeat must be greater than zero")
	}
	return count, i + 1, nil
}

func randomIndex(random io.Reader, size int) (int, error) {
	value, err := crand.Int(random, big.NewInt(int64(size)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}
