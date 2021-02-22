package aead

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKDFValue(t *testing.T) {
	GeneratedKey := KDF([]byte("Demo Key for KDF Value Test"), "Demo Path for KDF Value Test", "Demo Path for KDF Value Test2", "Demo Path for KDF Value Test3")
	fmt.Println(hex.EncodeToString(GeneratedKey))
	assert.Equal(t, "53e9d7e1bd7bd25022b71ead07d8a596efc8a845c7888652fd684b4903dc8892", hex.EncodeToString(GeneratedKey), "Should generate expected KDF Value")
}
