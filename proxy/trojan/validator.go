package trojan

import (
	"sync"

	"v2ray.com/core/common/protocol"
)

type Validator struct {
	users sync.Map
}

func (v *Validator) Add(u *protocol.MemoryUser) error {
	user := u.Account.(*MemoryAccount)
	v.users.Store(hexString(user.Key), u)
	return nil
}

func (v *Validator) Get(hash string) *protocol.MemoryUser {
	u, _ := v.users.Load(hash)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}
