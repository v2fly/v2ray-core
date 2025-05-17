package mirrorcrypto

import (
	"crypto/hkdf"
	"crypto/sha256"
	"hash"
)

func DeriveEncryptionKey(primaryKey, clientRandom, serverRandom []byte, tag string) ([]byte, []byte, error) {
	if len(primaryKey) != 32 {
		return nil, nil, newError("invalid primary key size: ", len(primaryKey))
	}
	if len(clientRandom) != 32 {
		return nil, nil, newError("invalid client random size: ", len(clientRandom))
	}
	if len(serverRandom) != 32 {
		return nil, nil, newError("invalid server random size: ", len(serverRandom))
	}

	// Concatenate the primary key, client random, and server random
	combined := append(primaryKey, clientRandom...)
	combined = append(combined, serverRandom...)

	encryptionKey, err := hkdf.Expand(func() hash.Hash {
		return sha256.New()
	}, combined, "v2ray-sp76YMKM-EkGrFUNL-rTJRJMkU:tlsmirror-encryption", 12)
	if err != nil {
		return nil, nil, newError("unable to derive encryption key").Base(err)
	}

	nonceMask, err := hkdf.Expand(func() hash.Hash {
		return sha256.New()
	}, combined, "v2ray-sp76YMKM-EkGrFUNL-rTJRJMkU:tlsmirror-noncemask", 12)

	if err != nil {
		return nil, nil, newError("unable to derive nonce mask").Base(err)
	}

	return encryptionKey, nonceMask, nil
}
