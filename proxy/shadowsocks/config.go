package shadowsocks

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"io"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/antireplay"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/crypto"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
)

// MemoryAccount is an account type converted from Account.
type MemoryAccount struct {
	Cipher Cipher
	Key    []byte

	replayFilter antireplay.GeneralizedReplayFilter

	ReducedIVEntropy bool
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(another protocol.Account) bool {
	if account, ok := another.(*MemoryAccount); ok {
		return bytes.Equal(a.Key, account.Key)
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

func createAesGcm(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	common.Must(err)
	gcm, err := cipher.NewGCM(block)
	common.Must(err)
	return gcm
}

func createChaCha20Poly1305(key []byte) cipher.AEAD {
	ChaChaPoly1305, err := chacha20poly1305.New(key)
	common.Must(err)
	return ChaChaPoly1305
}

func (a *Account) getCipher() (Cipher, error) {
	switch a.CipherType {
	case CipherType_AES_128_GCM:
		return &AEADCipher{
			KeyBytes:        16,
			IVBytes:         16,
			AEADAuthCreator: createAesGcm,
		}, nil
	case CipherType_AES_256_GCM:
		return &AEADCipher{
			KeyBytes:        32,
			IVBytes:         32,
			AEADAuthCreator: createAesGcm,
		}, nil
	case CipherType_CHACHA20_POLY1305:
		return &AEADCipher{
			KeyBytes:        32,
			IVBytes:         32,
			AEADAuthCreator: createChaCha20Poly1305,
		}, nil
	case CipherType_NONE:
		return NoneCipher{}, nil
	default:
		return nil, newError("Unsupported cipher.")
	}
}

// AsAccount implements protocol.AsAccount.
func (a *Account) AsAccount() (protocol.Account, error) {
	Cipher, err := a.getCipher()
	if err != nil {
		return nil, newError("failed to get cipher").Base(err)
	}
	return &MemoryAccount{
		Cipher: Cipher,
		Key:    passwordToCipherKey([]byte(a.Password), Cipher.KeySize()),
		replayFilter: func() antireplay.GeneralizedReplayFilter {
			if a.IvCheck {
				return antireplay.NewBloomRing()
			}
			return nil
		}(),
		ReducedIVEntropy: a.ExperimentReducedIvHeadEntropy,
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

type AEADCipher struct {
	KeyBytes        int32
	IVBytes         int32
	AEADAuthCreator func(key []byte) cipher.AEAD
}

func (*AEADCipher) IsAEAD() bool {
	return true
}

func (c *AEADCipher) KeySize() int32 {
	return c.KeyBytes
}

func (c *AEADCipher) IVSize() int32 {
	return c.IVBytes
}

func (c *AEADCipher) createAuthenticator(key []byte, iv []byte) *crypto.AEADAuthenticator {
	nonce := crypto.GenerateInitialAEADNonce()
	subkey := make([]byte, c.KeyBytes)
	hkdfSHA1(key, iv, subkey)
	return &crypto.AEADAuthenticator{
		AEAD:           c.AEADAuthCreator(subkey),
		NonceGenerator: nonce,
	}
}

func (c *AEADCipher) NewEncryptionWriter(key []byte, iv []byte, writer io.Writer) (buf.Writer, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationWriter(auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, writer, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) NewDecryptionReader(key []byte, iv []byte, reader io.Reader) (buf.Reader, error) {
	auth := c.createAuthenticator(key, iv)
	return crypto.NewAuthenticationReader(auth, &crypto.AEADChunkSizeParser{
		Auth: auth,
	}, reader, protocol.TransferTypeStream, nil), nil
}

func (c *AEADCipher) EncodePacket(key []byte, b *buf.Buffer) error {
	ivLen := c.IVSize()
	payloadLen := b.Len()
	auth := c.createAuthenticator(key, b.BytesTo(ivLen))

	b.Extend(int32(auth.Overhead()))
	_, err := auth.Seal(b.BytesTo(ivLen), b.BytesRange(ivLen, payloadLen))
	return err
}

func (c *AEADCipher) DecodePacket(key []byte, b *buf.Buffer) error {
	if b.Len() <= c.IVSize() {
		return newError("insufficient data: ", b.Len())
	}
	ivLen := c.IVSize()
	payloadLen := b.Len()
	auth := c.createAuthenticator(key, b.BytesTo(ivLen))

	bbb, err := auth.Open(b.BytesTo(ivLen), b.BytesRange(ivLen, payloadLen))
	if err != nil {
		return err
	}
	b.Resize(ivLen, int32(len(bbb)))
	return nil
}

type NoneCipher struct{}

func (NoneCipher) KeySize() int32 { return 0 }
func (NoneCipher) IVSize() int32  { return 0 }
func (NoneCipher) IsAEAD() bool {
	return false
}

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
	case "aes-128-gcm", "aes_128_gcm", "aead_aes_128_gcm":
		return CipherType_AES_128_GCM
	case "aes-256-gcm", "aes_256_gcm", "aead_aes_256_gcm":
		return CipherType_AES_256_GCM
	case "chacha20-poly1305", "chacha20_poly1305", "aead_chacha20_poly1305", "chacha20-ietf-poly1305":
		return CipherType_CHACHA20_POLY1305
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
	return key
}

func hkdfSHA1(secret, salt, outKey []byte) {
	r := hkdf.New(sha1.New, secret, salt, []byte("ss-subkey"))
	common.Must2(io.ReadFull(r, outKey))
}
