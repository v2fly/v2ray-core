package shadowsocksr

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "crypto/rc4"
    "crypto/sha1"
    "io"
    "strings"
    
    "golang.org/x/crypto/chacha20"
    "golang.org/x/crypto/hkdf"
    
    "github.com/v2fly/v2ray-core/v5/common"
    "github.com/v2fly/v2ray-core/v5/common/antireplay"
    "github.com/v2fly/v2ray-core/v5/common/buf"
    "github.com/v2fly/v2ray-core/v5/common/crypto"
    "github.com/v2fly/v2ray-core/v5/common/protocol"
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct {
    Cipher      Cipher
    Key         []byte
    
    Protocol    string
    ProtocolParam string
    Obfs       string
    ObfsParam  string
    
    replayFilter antireplay.GeneralizedReplayFilter
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(another protocol.Account) bool {
    if account, ok := another.(*MemoryAccount); ok {
        return bytes.Equal(a.Key, account.Key) &&
               a.Protocol == account.Protocol &&
               a.ProtocolParam == account.ProtocolParam &&
               a.Obfs == account.Obfs &&
               a.ObfsParam == account.ObfsParam
    }
    return false
}

func (a *MemoryAccount) CheckIV(iv []byte) error {
    if a.replayFilter == nil {
        return nil
    }
    if a.replayFilter.Check(iv) {
        return nil
    }
    return newError("IV is not unique")
}

func createAesStreamFunc(key []byte, iv []byte) (cipher.Stream, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    return cipher.NewCFBEncrypter(block, iv), nil
}

func createChaCha20Stream(key []byte, iv []byte) (cipher.Stream, error) {
    return chacha20.NewUnauthenticatedCipher(key, iv)
}

func createRC4MD5Stream(key []byte, iv []byte) (cipher.Stream, error) {
    h := md5.New()
    h.Write(key)
    h.Write(iv)
    rc4Key := h.Sum(nil)
    return rc4.NewCipher(rc4Key)
}

func (a *Account) getCipher() (Cipher, error) {
    switch a.CipherType {
    case CipherType_AES_128_CFB:
        return &StreamCipher{
            KeyBytes: 16,
            IVBytes: 16,
            NewStream: createAesStreamFunc,
        }, nil
    case CipherType_AES_192_CFB:
        return &StreamCipher{
            KeyBytes: 24,
            IVBytes: 16,
            NewStream: createAesStreamFunc,
        }, nil
    case CipherType_AES_256_CFB:
        return &StreamCipher{
            KeyBytes: 32,
            IVBytes: 16,
            NewStream: createAesStreamFunc,
        }, nil
    case CipherType_CHACHA20:
        return &StreamCipher{
            KeyBytes: 32,
            IVBytes: 8,
            NewStream: createChaCha20Stream,
        }, nil
    case CipherType_CHACHA20_IETF:
        return &StreamCipher{
            KeyBytes: 32,
            IVBytes: 12,
            NewStream: createChaCha20Stream,
        }, nil
    case CipherType_RC4_MD5:
        return &StreamCipher{
            KeyBytes: 16,
            IVBytes: 16,
            NewStream: createRC4MD5Stream,
        }, nil
    case CipherType_NONE:
        return NoneCipher{}, nil
    default:
        return nil, newError("Unsupported cipher.")
    }
}

// AsAccount implements protocol.AsAccount.
func (a *Account) AsAccount() (protocol.Account, error) {
    cipher, err := a.getCipher()
    if err != nil {
        return nil, newError("failed to get cipher").Base(err)
    }
    
    return &MemoryAccount{
        Cipher: cipher,
        Key:    passwordToCipherKey([]byte(a.Password), cipher.KeySize()),
        Protocol: a.Protocol,
        ProtocolParam: a.ProtocolParam,
        Obfs: a.Obfs,
        ObfsParam: a.ObfsParam,
        replayFilter: func() antireplay.GeneralizedReplayFilter {
            return antireplay.NewBloomRing()
        }(),
    }, nil
}

// Cipher is an interface for all Shadowsocks ciphers.
type Cipher interface {
    KeySize() int32
    IVSize() int32
    NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error)
    NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error)
    IsAEAD() bool
    EncodePacket(key []byte, b *buf.Buffer) error
    DecodePacket(key []byte, b *buf.Buffer) error
}

// StreamCipher implements all stream ciphers
type StreamCipher struct {
    KeyBytes  int32
    IVBytes   int32
    NewStream func(key []byte, iv []byte) (cipher.Stream, error)
}

func (c *StreamCipher) KeySize() int32 {
    return c.KeyBytes
}

func (c *StreamCipher) IVSize() int32 {
    return c.IVBytes
}

func (c *StreamCipher) IsAEAD() bool {
    return false
}

func (c *StreamCipher) NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error) {
    stream, err := c.NewStream(key, iv)
    if err != nil {
        return nil, err
    }
    return &buf.SequentialWriter{Writer: crypto.NewStreamWriter(stream, writer)}, nil
}

func (c *StreamCipher) NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error) {
    stream, err := c.NewStream(key, iv)
    if err != nil {
        return nil, err
    }
    return &buf.SequentialReader{Reader: crypto.NewStreamReader(stream, reader)}, nil
}

func (c *StreamCipher) EncodePacket(key []byte, b *buf.Buffer) error {
    ivLen := c.IVSize()
    payloadLen := b.Len()
    stream, err := c.NewStream(key, b.BytesTo(ivLen))
    if err != nil {
        return err
    }

    stream.XORKeyStream(b.BytesFrom(ivLen), b.BytesRange(ivLen, payloadLen))
    return nil
}

func (c *StreamCipher) DecodePacket(key []byte, b *buf.Buffer) error {
    if b.Len() <= c.IVSize() {
        return newError("insufficient data: ", b.Len())
    }
    ivLen := c.IVSize()
    payloadLen := b.Len()
    stream, err := c.NewStream(key, b.BytesTo(ivLen))
    if err != nil {
        return err
    }

    stream.XORKeyStream(b.BytesFrom(ivLen), b.BytesRange(ivLen, payloadLen))
    return nil
}

type NoneCipher struct{}

func (NoneCipher) KeySize() int32 { return 0 }
func (NoneCipher) IVSize() int32  { return 0 }
func (NoneCipher) IsAEAD() bool   { return false }

func (NoneCipher) NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error) {
    return buf.NewReader(reader), nil
}

func (NoneCipher) NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error) {
    return buf.NewWriter(writer), nil
}

func (NoneCipher) EncodePacket(key []byte, b *buf.Buffer) error {
    return nil
}

func (NoneCipher) DecodePacket(key []byte, b *buf.Buffer) error {
    return nil
}

func CipherFromString(c string) CipherType {
    switch strings.ToLower(c) {
    case "aes-128-cfb":
        return CipherType_AES_128_CFB
    case "aes-192-cfb":
        return CipherType_AES_192_CFB
    case "aes-256-cfb":
        return CipherType_AES_256_CFB
    case "chacha20":
        return CipherType_CHACHA20
    case "chacha20-ietf":
        return CipherType_CHACHA20_IETF
    case "rc4-md5":
        return CipherType_RC4_MD5
    case "none", "plain":
        return CipherType_NONE
    default:
        return CipherType_UNKNOWN
    }
}

func passwordToCipherKey(password []byte, keySize int32) []byte {
    key := make([]byte, 0, keySize)

    md5Sum := md5.Sum(password)
    key = append(key, md5Sum[:]...)

    for int32(len(key)) < keySize {
        md5Hash := md5.New()
        common.Must2(md5Hash.Write(md5Sum[:]))
        common.Must2(md5Hash.Write(password))
        md5Hash.Sum(md5Sum[:0])

        key = append(key, md5Sum[:]...)
    }
    return key[:keySize]
}

func hkdfSHA1(secret, salt, outKey []byte) {
    r := hkdf.New(sha1.New, secret, salt, []byte("ss-subkey"))
    common.Must2(io.ReadFull(r, outKey))
}
