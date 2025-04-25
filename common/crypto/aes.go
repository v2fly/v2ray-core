package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/ghxhy/v2ray-core/v5/common"
)

// NewAesDecryptionStream creates a new AES encryption stream based on given key and IV.
// Caller must ensure the length of key and IV is either 16, 24 or 32 bytes.
func NewAesDecryptionStream(key []byte, iv []byte) cipher.Stream {
	return NewAesStreamMethod(key, iv, cipher.NewCFBDecrypter)
}

// NewAesEncryptionStream creates a new AES description stream based on given key and IV.
// Caller must ensure the length of key and IV is either 16, 24 or 32 bytes.
func NewAesEncryptionStream(key []byte, iv []byte) cipher.Stream {
	return NewAesStreamMethod(key, iv, cipher.NewCFBEncrypter)
}

func NewAesStreamMethod(key []byte, iv []byte, f func(cipher.Block, []byte) cipher.Stream) cipher.Stream {
	aesBlock, err := aes.NewCipher(key)
	common.Must(err)
	return f(aesBlock, iv)
}

// NewAesCTRStream creates a stream cipher based on AES-CTR.
func NewAesCTRStream(key []byte, iv []byte) cipher.Stream {
	return NewAesStreamMethod(key, iv, cipher.NewCTR)
}

// NewAesGcm creates a AEAD cipher based on AES-GCM.
func NewAesGcm(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	common.Must(err)
	aead, err := cipher.NewGCM(block)
	common.Must(err)
	return aead
}
