package mirrorcrypto

import (
	"crypto/cipher"

	"github.com/v2fly/v2ray-core/v5/common/crypto"
)

type Decryptor struct {
	nonceGenerator crypto.BytesGenerator
	aead           cipher.AEAD
	nextNonce      []byte
}

func NewDecryptor(encryptionKey []byte, nonceMask []byte) *Decryptor {
	wrappedAead := aeadAESGCMTLS13(encryptionKey, nonceMask)
	return &Decryptor{
		nonceGenerator: generateInitialAEADNonce(),
		aead:           wrappedAead,
	}
}

func (d *Decryptor) Open(dst, src []byte) ([]byte, error) {
	if d.nextNonce == nil {
		d.nextNonce = d.nonceGenerator()
	}
	dst, err := d.aead.Open(dst[:0], d.nextNonce, src, nil)
	if err != nil {
		return nil, err
	}
	d.nextNonce = nil
	return dst, nil
}

func (d *Decryptor) NonceSize() int {
	return d.aead.NonceSize()
}
