package mirrorcrypto

import (
	"crypto/hkdf"
	"crypto/sha256"
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
	combined := append(primaryKey, clientRandom...) // nolint: gocritic
	combined = append(combined, serverRandom...)

	encryptionKey, err := hkdf.Expand(sha256.New, combined, "v2ray-sp76YMKM-EkGrFUNL-rTJRJMkU:tlsmirror-encryption"+tag, 16)
	if err != nil {
		return nil, nil, newError("unable to derive encryption key").Base(err)
	}

	nonceMask, err := hkdf.Expand(sha256.New, combined, "v2ray-sp76YMKM-EkGrFUNL-rTJRJMkU:tlsmirror-noncemask"+tag, 12)
	if err != nil {
		return nil, nil, newError("unable to derive nonce mask").Base(err)
	}

	return encryptionKey, nonceMask, nil
}

func DeriveSecondaryKey(primaryKey []byte, tag string) ([]byte, error) {
	if len(primaryKey) != 32 {
		return nil, newError("invalid primary key size: ", len(primaryKey))
	}

	// Use HKDF to derive a secondary key
	secondaryKey, err := hkdf.Expand(sha256.New, primaryKey, "v2ray-sv77RCEY-e8AhYsbD-BmFC7XRK:tlsmirror-secondary"+tag, 16)
	if err != nil {
		return nil, newError("unable to derive secondary key").Base(err)
	}

	return secondaryKey, nil
}

func DeriveSequenceWatermarkingKey(primaryKey, clientRandom, serverRandom []byte, tag string) ([]byte, []byte, error) {
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
	combined := append(primaryKey, clientRandom...) // nolint: gocritic
	combined = append(combined, serverRandom...)

	encryptionKey, err := hkdf.Expand(sha256.New, combined, "v2ray-xv64FXUU-GxMn8UYz-bTy6UDeE:tlsmirror-sequence-watermark"+tag, 32)
	if err != nil {
		return nil, nil, newError("unable to derive encryption key").Base(err)
	}

	nonceMask, err := hkdf.Expand(sha256.New, combined, "v2ray-xv64FXUU-GxMn8UYz-bTy6UDeE:tlsmirror-sequence-watermark"+tag, 24)
	if err != nil {
		return nil, nil, newError("unable to derive nonce mask").Base(err)
	}

	return encryptionKey, nonceMask, nil
}
