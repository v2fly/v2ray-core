package shadowsocks2022

import (
	"crypto/aes"
	"crypto/cipher"
)

func newAES128GCMMethod() *AES128GCMMethod {
	return &AES128GCMMethod{}
}

type AES128GCMMethod struct{}

func (a AES128GCMMethod) GetSessionSubKeyAndSaltLength() int {
	return 16
}

func (a AES128GCMMethod) GetStreamAEAD(sessionSubKey []byte) (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(sessionSubKey)
	if err != nil {
		return nil, newError("failed to create AES cipher").Base(err)
	}
	aead, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, newError("failed to create AES-GCM AEAD").Base(err)
	}
	return aead, nil
}

func (a AES128GCMMethod) GenerateEIH(currentIdentitySubKey []byte, nextPskHash []byte, out []byte) error {
	aesCipher, err := aes.NewCipher(currentIdentitySubKey)
	if err != nil {
		return newError("failed to create AES cipher").Base(err)
	}
	aesCipher.Encrypt(out, nextPskHash)
	return nil
}

func (a AES128GCMMethod) GetUDPClientProcessor(ipsk [][]byte, psk []byte, derivation KeyDerivation) (UDPClientPacketProcessor, error) {
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
		sessionKey := make([]byte, a.GetSessionSubKeyAndSaltLength())
		derivation.GetSessionSubKey(psk, sessionID, sessionKey)
		block, err := aes.NewCipher(sessionKey)
		if err != nil {
			panic(err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			panic(err)
		}
		return aead
	}
	if len(ipsk) == 0 {
		return NewAESUDPClientPacketProcessor(reqSeparateHeaderCipher, respSeparateHeaderCipher, getPacketAEAD, nil), nil
	}
	eihGenerator := newAESEIHGeneratorContainer(len(ipsk), psk, ipsk)
	getEIH := func(mask []byte) ExtensibleIdentityHeaders {
		eih, err := eihGenerator.GenerateEIHUDP(derivation, a, mask)
		if err != nil {
			newError("failed to generate EIH").Base(err).WriteToLog()
		}
		return eih
	}
	return NewAESUDPClientPacketProcessor(reqSeparateHeaderCipher, respSeparateHeaderCipher, getPacketAEAD, getEIH), nil
}
