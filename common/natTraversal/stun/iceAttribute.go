package stun

import (
	"github.com/pion/stun/v3"
	stunv4 "github.com/pion/stun/v4"
)

// iceAttribute adapts ICE's STUN v4 attributes while preserving this package's
// public STUN v3 message and setter types.
type iceAttribute struct {
	setter stunv4.Setter
}

func (a iceAttribute) AddTo(message *stun.Message) error {
	encoded, err := stunv4.Build(a.setter)
	if err != nil {
		return err
	}
	for _, attribute := range encoded.Attributes {
		message.Add(stun.AttrType(attribute.Type), attribute.Value)
	}
	return nil
}
