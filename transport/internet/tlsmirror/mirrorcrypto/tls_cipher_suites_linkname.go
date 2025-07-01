package mirrorcrypto

import _ "unsafe"
import "crypto/cipher"

// This linkname is necessary to avoid duplicating too many internal packages.

//go:linkname aeadAESGCMTLS13 crypto/tls.aeadAESGCMTLS13
func aeadAESGCMTLS13(key, nonceMask []byte) cipher.AEAD
