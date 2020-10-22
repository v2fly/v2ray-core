package trojan

import (
	"strings"
	"sync"

	"v2ray.com/core/common/protocol"
)

// Validator stores valid trojan users
type Validator struct {
	// Considering email's usage here, map + sync.Mutex/RWMutex may have better performance.
	email sync.Map
	users sync.Map
}

// Add a trojan user
func (v *Validator) Add(u *protocol.MemoryUser) error {
	if u.Email != "" {
		_, loaded := v.email.LoadOrStore(strings.ToLower(u.Email), u)
		if loaded {
			return newError("User ", u.Email, " already exists.")
		}
	}

	account := u.Account.(*MemoryAccount)
	v.users.Store(hexString(account.Key), u)
	return nil
}

// Del a trojan user
func (v *Validator) Del(e string) error {
	if e == "" {
		return newError("Email must not be empty.")
	}
	le := strings.ToLower(e)
	u, _ := v.email.Load(le)
	if u == nil {
		return newError("User ", e, " not found.")
	}
	v.email.Delete(le)

	account := u.(*protocol.MemoryUser).Account.(*MemoryAccount)
	v.users.Delete(hexString(account.Key))
	return nil
}

// Get user with hashed key, nil if user doesn't exist.
func (v *Validator) Get(hash string) *protocol.MemoryUser {
	u, _ := v.users.Load(hash)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}
