package aead

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
)

func KDF(key []byte, path ...string) []byte {
	var hmacf hash.Hash
	for _, v := range path {
		hmacf = hmac.New(func() hash.Hash {
			return hmac.New(sha256.New, []byte(KDFSaltConstVMessAEADKDF))
		}, []byte(v))
	}
	hmacf.Write(key)
	return hmacf.Sum(nil)
}

func KDF16(key []byte, path ...string) []byte {
	r := KDF(key, path...)
	return r[:16]
}
