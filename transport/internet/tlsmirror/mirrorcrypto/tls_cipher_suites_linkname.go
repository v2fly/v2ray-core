package mirrorcrypto

import (
	"crypto/cipher"
	_ "unsafe"
)

// This linkname is necessary to avoid duplicating too many internal packages.

//go:linkname aeadAESGCMTLS13 crypto/tls.aeadAESGCMTLS13
func aeadAESGCMTLS13(key, nonceMask []byte) cipher.AEAD
