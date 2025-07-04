package mirrorbase

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/common/crypto"
)

func TestTLS12ExplicitNonceGeneration(t *testing.T) {
	generator := reverseBytesGeneratorByteOrder(crypto.GenerateIncreasingNonce([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}))

	firstValue := generator()
	if diff := cmp.Diff(firstValue, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}); diff != "" {
		t.Errorf("Unexpected first value: %s", diff)
	}

	secondValue := generator()
	if diff := cmp.Diff(secondValue, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}); diff != "" {
		t.Errorf("Unexpected second value: %s", diff)
	}

	thirdValue := generator()
	if diff := cmp.Diff(thirdValue, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03}); diff != "" {
		t.Errorf("Unexpected third value: %s", diff)
	}
}
