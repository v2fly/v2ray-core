package shadowsocks2022

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAES128GCMMethod() *AES128GCMMethod {
	return &AES128GCMMethod{}
}

type AES128GCMMethod struct {
}

func (A AES128GCMMethod) GetSessionSubKeyAndSaltLength() int {
	return 16
}

func (A AES128GCMMethod) GetStreamAEAD(SessionSubKey []byte) (cipher.AEAD, error) {
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

func (A AES128GCMMethod) GenerateEIH(CurrentIdentitySubKey []byte, nextPskHash []byte, out []byte) error {
	aesCipher, err := aes.NewCipher(CurrentIdentitySubKey)
	if err != nil {
		return newError("failed to create AES cipher").Base(err)
	}
	aesCipher.Encrypt(out, nextPskHash)
	return nil
}
