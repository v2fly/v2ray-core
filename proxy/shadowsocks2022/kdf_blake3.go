package shadowsocks2022

import (
	"lukechampine.com/blake3"

	"github.com/v2fly/v2ray-core/v5/common/buf"
)

func newBLAKE3KeyDerivation() *BLAKE3KeyDerivation {
	return &BLAKE3KeyDerivation{}
}

type BLAKE3KeyDerivation struct{}

func (b BLAKE3KeyDerivation) GetSessionSubKey(effectivePsk, salt []byte, outKey []byte) error {
	keyingMaterialBuffer := buf.New()
	keyingMaterialBuffer.Write(effectivePsk)
	keyingMaterialBuffer.Write(salt)
	blake3.DeriveKey(outKey, "shadowsocks 2022 session subkey", keyingMaterialBuffer.Bytes())
	keyingMaterialBuffer.Release()
	return nil
}

func (b BLAKE3KeyDerivation) GetIdentitySubKey(effectivePsk, salt []byte, outKey []byte) error {
	keyingMaterialBuffer := buf.New()
	keyingMaterialBuffer.Write(effectivePsk)
	keyingMaterialBuffer.Write(salt)
	blake3.DeriveKey(outKey, "shadowsocks 2022 identity subkey", keyingMaterialBuffer.Bytes())
	keyingMaterialBuffer.Release()
	return nil
}
