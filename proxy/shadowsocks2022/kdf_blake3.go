package shadowsocks2022

import (
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"lukechampine.com/blake3"
)

func newBLAKE3KeyDerivation() *BLAKE3KeyDerivation {
	return &BLAKE3KeyDerivation{}
}

type BLAKE3KeyDerivation struct {
}

func (B BLAKE3KeyDerivation) GetSessionSubKey(effectivePsk, Salt []byte, OutKey []byte) error {
	keyingMaterialBuffer := buf.New()
	keyingMaterialBuffer.Write(effectivePsk)
	keyingMaterialBuffer.Write(Salt)
	blake3.DeriveKey(OutKey, "shadowsocks 2022 session subkey", keyingMaterialBuffer.Bytes())
	keyingMaterialBuffer.Release()
	return nil
}

func (B BLAKE3KeyDerivation) GetIdentitySubKey(effectivePsk, Salt []byte, OutKey []byte) error {
	keyingMaterialBuffer := buf.New()
	keyingMaterialBuffer.Write(effectivePsk)
	keyingMaterialBuffer.Write(Salt)
	blake3.DeriveKey(OutKey, "shadowsocks 2022 identity subkey", keyingMaterialBuffer.Bytes())
	keyingMaterialBuffer.Release()
	return nil
}
