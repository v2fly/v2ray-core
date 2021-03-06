package templates

import "encoding/json"

type UniversalDot struct {
	Content []byte

	AsJson interface{}
}

func NewUniversalDot(content []byte) *UniversalDot {
	return &UniversalDot{Content: content}
}

func (ud *UniversalDot) IsJson() bool {
	if nil == json.Unmarshal(ud.Content, &ud.AsJson) {
		return true
	}
	return false
}
