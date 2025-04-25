package encoding

import (
	"crypto/md5"
	"encoding/binary"
	"hash/fnv"

	"golang.org/x/crypto/sha3"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/crypto"
)

// Authenticate authenticates a byte array using Fnv hash.
func Authenticate(b []byte) uint32 {
	fnv1hash := fnv.New32a()
	common.Must2(fnv1hash.Write(b))
	return fnv1hash.Sum32()
}

type NoOpAuthenticator struct{}

func (NoOpAuthenticator) NonceSize() int {
	return 0
}

func (NoOpAuthenticator) Overhead() int {
	return 0
}

// Seal implements AEAD.Seal().
func (NoOpAuthenticator) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	return append(dst[:0], plaintext...)
}

// Open implements AEAD.Open().
func (NoOpAuthenticator) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	return append(dst[:0], ciphertext...), nil
}

// FnvAuthenticator is an AEAD based on Fnv hash.
type FnvAuthenticator struct{}

// NonceSize implements AEAD.NonceSize().
func (*FnvAuthenticator) NonceSize() int {
	return 0
}

// Overhead impelements AEAD.Overhead().
func (*FnvAuthenticator) Overhead() int {
	return 4
}

// Seal implements AEAD.Seal().
func (*FnvAuthenticator) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	dst = append(dst, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(dst, Authenticate(plaintext))
	return append(dst, plaintext...)
}

// Open implements AEAD.Open().
func (*FnvAuthenticator) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if binary.BigEndian.Uint32(ciphertext[:4]) != Authenticate(ciphertext[4:]) {
		return dst, newError("invalid authentication")
	}
	return append(dst, ciphertext[4:]...), nil
}

// GenerateChacha20Poly1305Key generates a 32-byte key from a given 16-byte array.
func GenerateChacha20Poly1305Key(b []byte) []byte {
	key := make([]byte, 32)
	t := md5.Sum(b)
	copy(key, t[:])
	t = md5.Sum(key[:16])
	copy(key[16:], t[:])
	return key
}

type ShakeSizeParser struct {
	shake  sha3.ShakeHash
	buffer [2]byte
}

func NewShakeSizeParser(nonce []byte) *ShakeSizeParser {
	shake := sha3.NewShake128()
	common.Must2(shake.Write(nonce))
	return &ShakeSizeParser{
		shake: shake,
	}
}

func (*ShakeSizeParser) SizeBytes() int32 {
	return 2
}

func (s *ShakeSizeParser) next() uint16 {
	common.Must2(s.shake.Read(s.buffer[:]))
	return binary.BigEndian.Uint16(s.buffer[:])
}

func (s *ShakeSizeParser) Decode(b []byte) (uint16, error) {
	mask := s.next()
	size := binary.BigEndian.Uint16(b)
	return mask ^ size, nil
}

func (s *ShakeSizeParser) Encode(size uint16, b []byte) []byte {
	mask := s.next()
	binary.BigEndian.PutUint16(b, mask^size)
	return b[:2]
}

func (s *ShakeSizeParser) NextPaddingLen() uint16 {
	return s.next() % 64
}

func (s *ShakeSizeParser) MaxPaddingLen() uint16 {
	return 64
}

type AEADSizeParser struct {
	crypto.AEADChunkSizeParser
}

func NewAEADSizeParser(auth *crypto.AEADAuthenticator) *AEADSizeParser {
	return &AEADSizeParser{crypto.AEADChunkSizeParser{Auth: auth}}
}
