package hysteria2

import (
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct{}

// AsAccount implements protocol.AsAccount.
func (a *Account) AsAccount() (protocol.Account, error) {
	return &MemoryAccount{}, nil
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(another protocol.Account) bool {
	return false
}
