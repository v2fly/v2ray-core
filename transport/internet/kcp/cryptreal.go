package kcp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"github.com/v2fly/v2ray-core/common"
)

func NewAEADAESGCMBasedOnSeed(seed string) cipher.AEAD {
	HashedSeed := sha256.Sum256([]byte(seed))
	aesBlock := common.Must2(aes.NewCipher(HashedSeed[:16])).(cipher.Block)
	return common.Must2(cipher.NewGCM(aesBlock)).(cipher.AEAD)
}
