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

func (A AES256GCMMethod) GetUDPClientProcessor(ipsk [][]byte, psk []byte, derivation KeyDerivation) (UDPClientPacketProcessor, error) {
	reqSeparateHeaderPsk := psk
	if ipsk != nil {
		reqSeparateHeaderPsk = ipsk[0]
	}
	reqSeparateHeaderCipher, err := aes.NewCipher(reqSeparateHeaderPsk)
	if err != nil {
		return nil, newError("failed to create AES cipher").Base(err)
	}
	respSeparateHeaderCipher, err := aes.NewCipher(psk)
	if err != nil {
		return nil, newError("failed to create AES cipher").Base(err)
	}
	getPacketAEAD := func(sessionID []byte) cipher.AEAD {
		sessionKey := make([]byte, A.GetSessionSubKeyAndSaltLength())
		derivation.GetSessionSubKey(psk, sessionID, sessionKey)
		block, err := aes.NewCipher(sessionKey)
		aead, err := cipher.NewGCM(block)
		if err != nil {
			panic(err)
		}
		return aead
	}
	eihGenerator := newAESEIHGeneratorContainer(len(ipsk), psk, ipsk)
	getEIH := func(mask []byte) ExtensibleIdentityHeaders {
		eih, err := eihGenerator.GenerateEIHUDP(derivation, A, mask)
		if err != nil {
			newError("failed to generate EIH").Base(err).WriteToLog()
		}
		return eih
	}
	return NewAESUDPClientPacketProcessor(reqSeparateHeaderCipher, respSeparateHeaderCipher, getPacketAEAD, getEIH), nil
}
