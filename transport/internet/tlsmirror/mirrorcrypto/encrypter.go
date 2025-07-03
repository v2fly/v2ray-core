package mirrorcrypto

import (
	"crypto/cipher"

	"github.com/v2fly/v2ray-core/v5/common/crypto"
)

type Encryptor struct {
	nonceGenerator crypto.BytesGenerator
	aead           cipher.AEAD
}

func NewEncryptor(encryptionKey []byte, nonceMask []byte) *Encryptor {
	wrappedAead := aeadAESGCMTLS13(encryptionKey, nonceMask)
	return &Encryptor{
		nonceGenerator: generateInitialAEADNonce(),
		aead:           wrappedAead,
	}
}

func (e *Encryptor) Seal(dst, src []byte) ([]byte, error) {
	nonce := e.nonceGenerator()
	dst = e.aead.Seal(dst, nonce, src, nil)
	return dst, nil
}

func (e *Encryptor) NonceSize() int {
	return e.aead.NonceSize()
}
