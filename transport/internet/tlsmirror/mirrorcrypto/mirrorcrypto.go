package mirrorcrypto

import "github.com/v2fly/v2ray-core/v5/common/crypto"

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func generateInitialAEADNonce() crypto.BytesGenerator {
	return crypto.GenerateIncreasingNonce([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
}
