package shadowsocksr

import (
    "github.com/v2fly/v2ray-core/v5/common/protocol"
)

type CipherType int32

const (
    CipherType_UNKNOWN        CipherType = 0
    CipherType_NONE          CipherType = 1
    CipherType_AES_128_CFB   CipherType = 2
    CipherType_AES_192_CFB   CipherType = 3
    CipherType_AES_256_CFB   CipherType = 4
    CipherType_AES_128_CTR   CipherType = 5
    CipherType_AES_192_CTR   CipherType = 6
    CipherType_AES_256_CTR   CipherType = 7
    CipherType_RC4_MD5       CipherType = 8
    CipherType_CHACHA20      CipherType = 9
    CipherType_CHACHA20_IETF CipherType = 10
)

// Account represents a Shadowsocksr account
type Account struct {
    Password       string     `json:"password"`
    CipherType     CipherType `json:"cipher_type"`
    Protocol       string     `json:"protocol"`
    ProtocolParam  string     `json:"protocol_param"`
    Obfs           string     `json:"obfs"`
    ObfsParam      string     `json:"obfs_param"`
}

// ServerConfig represents a Shadowsocksr server configuration
type ServerConfig struct {
    User    *protocol.User        `json:"user"`
    Network []protocol.Network    `json:"network"`
}

// ClientConfig represents a Shadowsocksr client configuration
type ClientConfig struct {
    Server []*protocol.ServerEndpoint `json:"server"`
}

// GetCipherName converts cipher type to string
func (c CipherType) GetCipherName() string {
    switch c {
    case CipherType_NONE:
        return "NONE"
    case CipherType_AES_128_CFB:
        return "AES-128-CFB"
    case CipherType_AES_192_CFB:
        return "AES-192-CFB"
    case CipherType_AES_256_CFB:
        return "AES-256-CFB"
    case CipherType_AES_128_CTR:
        return "AES-128-CTR"
    case CipherType_AES_192_CTR:
        return "AES-192-CTR"
    case CipherType_AES_256_CTR:
        return "AES-256-CTR"
    case CipherType_RC4_MD5:
        return "RC4-MD5"
    case CipherType_CHACHA20:
        return "CHACHA20"
    case CipherType_CHACHA20_IETF:
        return "CHACHA20-IETF"
    default:
        return "UNKNOWN"
    }
}

// AsAccount implements protocol.Account interface
func (a *Account) AsAccount() (protocol.Account, error) {
    return &MemoryAccount{
        Password:   a.Password,
        CipherType: a.CipherType,
        Protocol:  a.Protocol,
        ProtocolParam: a.ProtocolParam,
        Obfs:     a.Obfs,
        ObfsParam: a.ObfsParam,
    }, nil
}

// Equals implements protocol.Account interface
func (a *Account) Equals(another protocol.Account) bool {
    if account, ok := another.(*Account); ok {
        return a.Password == account.Password &&
            a.CipherType == account.CipherType &&
            a.Protocol == account.Protocol &&
            a.ProtocolParam == account.ProtocolParam &&
            a.Obfs == account.Obfs &&
            a.ObfsParam == account.ObfsParam
    }
    return false
}

// MemoryAccount is an in-memory form of Account
type MemoryAccount struct {
    *Account
    Key     []byte
    Cipher  interface{}
}
