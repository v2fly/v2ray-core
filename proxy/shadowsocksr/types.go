package shadowsocksr

import (
	"crypto/cipher"
	"encoding/json"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common/net"
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
	Password      string     `json:"password"`
	CipherType    CipherType `json:"cipher_type"`
	Protocol      string     `json:"protocol"`
	ProtocolParam string     `json:"protocol_param"`
	Obfs          string     `json:"obfs"`
	ObfsParam     string     `json:"obfs_param"`
}

// MemoryAccount is an in-memory form of Account
type MemoryAccount struct {
	Password      string
	Key           []byte
	CipherType    CipherType
	Protocol      string
	ProtocolParam string
	Obfs          string
	ObfsParam     string
	Cipher        interface{}
}

// ServerConfig represents a Shadowsocksr server configuration
type ServerConfig struct {
	User    *protocol.User `json:"user"`
	Network []net.Network  `json:"network"`
}

// ClientConfig represents a Shadowsocksr client configuration
type ClientConfig struct {
	Server []*protocol.ServerEndpoint `json:"server"`
}

// GetCipherName converts cipher type to string
func (c CipherType) GetCipherName() string {
	switch c {
	case CipherType_NONE:
		return "none"
	case CipherType_AES_128_CFB:
		return "aes-128-cfb"
	case CipherType_AES_192_CFB:
		return "aes-192-cfb"
	case CipherType_AES_256_CFB:
		return "aes-256-cfb"
	case CipherType_AES_128_CTR:
		return "aes-128-ctr"
	case CipherType_AES_192_CTR:
		return "aes-192-ctr"
	case CipherType_AES_256_CTR:
		return "aes-256-ctr"
	case CipherType_RC4_MD5:
		return "rc4-md5"
	case CipherType_CHACHA20:
		return "chacha20"
	case CipherType_CHACHA20_IETF:
		return "chacha20-ietf"
	default:
		return "unknown"
	}
}

// CipherFromString converts string to cipher type
func CipherFromString(name string) CipherType {
	switch strings.ToLower(name) {
	case "none":
		return CipherType_NONE
	case "aes-128-cfb":
		return CipherType_AES_128_CFB
	case "aes-192-cfb":
		return CipherType_AES_192_CFB
	case "aes-256-cfb":
		return CipherType_AES_256_CFB
	case "aes-128-ctr":
		return CipherType_AES_128_CTR
	case "aes-192-ctr":
		return CipherType_AES_192_CTR
	case "aes-256-ctr":
		return CipherType_AES_256_CTR
	case "rc4-md5":
		return CipherType_RC4_MD5
	case "chacha20":
		return CipherType_CHACHA20
	case "chacha20-ietf":
		return CipherType_CHACHA20_IETF
	default:
		return CipherType_UNKNOWN
	}
}

// Cipher interface for Shadowsocksr encryption
type Cipher interface {
	KeySize() int
	IVSize() int
	NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error)
	NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error)
	IsAEAD() bool
	EncodePacket(key []byte, b *buf.Buffer) error
	DecodePacket(key []byte, b *buf.Buffer) error
}

// AsAccount implements protocol.Account interface
func (a *Account) AsAccount() (protocol.Account, error) {
	return &MemoryAccount{
		Password:      a.Password,
		CipherType:    a.CipherType,
		Protocol:      a.Protocol,
		ProtocolParam: a.ProtocolParam,
		Obfs:         a.Obfs,
		ObfsParam:    a.ObfsParam,
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

// MarshalJSON implements json.Marshaler
func (c CipherType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.GetCipherName())
}

// UnmarshalJSON implements json.Unmarshaler
func (c *CipherType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	*c = CipherFromString(name)
	return nil
}

// StreamConnContext represents a connection context for stream ciphers
type StreamConnContext struct {
	Cipher     cipher.Stream
	ServerKey  []byte
	UserKey    []byte
	IV         []byte
	ReadCount  uint64
	WriteCount uint64
}

// HashedPassword represents a hashed password
type HashedPassword struct {
	Password      string
	PasswordHash  []byte
	PasswordKeyIV []byte
}

func init() {
	protocol.RegisterAccount((*Account)(nil))
}
