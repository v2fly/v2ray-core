package shadowsocks2022

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAES256GCMMethod() *AES256GCMMethod {
	return &AES256GCMMethod{}
}

type AES256GCMMethod struct {
}

func (A AES256GCMMethod) GetSessionSubKeyAndSaltLength() int {
	return 32
}

func (A AES256GCMMethod) GetStreamAEAD(SessionSubKey []byte) (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(SessionSubKey)
	if err != nil {
		return nil, newError("failed to create AES cipher").Base(err)
	}
	aead, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, newError("failed to create AES-GCM AEAD").Base(err)
	}
	return aead, nil
}

func (A AES256GCMMethod) GenerateEIH(CurrentIdentitySubKey []byte, nextPskHash []byte, out []byte) error {
	aesCipher, err := aes.NewCipher(CurrentIdentitySubKey)
	if err != nil {
		return newError("failed to create AES cipher").Base(err)
	}
	aesCipher.Encrypt(out, nextPskHash)
	return nil
}
