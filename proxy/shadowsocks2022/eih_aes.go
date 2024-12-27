package shadowsocks2022

import (
	"crypto/subtle"
	"io"

	"github.com/v2fly/struc"

	"github.com/v2fly/v2ray-core/v5/common/buf"

	"lukechampine.com/blake3"
)

func newAESEIH(size int) *aesEIH {
	return &aesEIH{length: size}
}

func newAESEIHWithData(size int, eih [][aesEIHSize]byte) *aesEIH {
	return &aesEIH{length: size, eih: eih}
}

const aesEIHSize = 16

type aesEIH struct {
	eih    [][aesEIHSize]byte
	length int
}

func (a *aesEIH) Pack(p []byte, opt *struc.Options) (int, error) {
	var totalCopy int
	for i := 0; i < a.length; i++ {
		n := copy(p[aesEIHSize*i:aesEIHSize*(i+1)], a.eih[i][:])
		if n != 16 {
			return 0, newError("failed to pack aesEIH")
		}
		totalCopy += n
	}
	return totalCopy, nil
}

func (a *aesEIH) Unpack(r io.Reader, length int, opt *struc.Options) error {
	a.eih = make([][aesEIHSize]byte, a.length)
	for i := 0; i < a.length; i++ {
		n, err := r.Read(a.eih[i][:])
		if err != nil {
			return newError("failed to unpack aesEIH").Base(err)
		}
		if n != aesEIHSize {
			return newError("failed to unpack aesEIH")
		}
	}
	return nil
}

func (a *aesEIH) Size(opt *struc.Options) int {
	return a.length * aesEIHSize
}

func (a *aesEIH) String() string {
	return ""
}

const aesEIHPskHashSize = 16

type aesEIHGenerator struct {
	ipsk     [][]byte
	ipskHash [][aesEIHPskHashSize]byte
	psk      []byte
	pskHash  [aesEIHPskHashSize]byte
	length   int
}

func newAESEIHGeneratorContainer(size int, effectivePsk []byte, ipsk [][]byte) *aesEIHGenerator {
	var ipskHash [][aesEIHPskHashSize]byte
	for _, v := range ipsk {
		hash := blake3.Sum512(v)
		ipskHash = append(ipskHash, [aesEIHPskHashSize]byte(hash[:16]))
	}
	pskHashFull := blake3.Sum512(effectivePsk)
	pskHash := [aesEIHPskHashSize]byte(pskHashFull[:16])
	return &aesEIHGenerator{length: size, ipsk: ipsk, ipskHash: ipskHash, psk: effectivePsk, pskHash: pskHash}
}

func (a *aesEIHGenerator) GenerateEIH(derivation KeyDerivation, method Method, salt []byte) (ExtensibleIdentityHeaders, error) {
	return a.generateEIHWithMask(derivation, method, salt, nil)
}

func (a *aesEIHGenerator) GenerateEIHUDP(derivation KeyDerivation, method Method, mask []byte) (ExtensibleIdentityHeaders, error) {
	return a.generateEIHWithMask(derivation, method, nil, mask)
}

func (a *aesEIHGenerator) generateEIHWithMask(derivation KeyDerivation, method Method, salt, mask []byte) (ExtensibleIdentityHeaders, error) {
	eih := make([][16]byte, a.length)
	current := a.length - 1
	currentPskHash := a.pskHash
	for {
		identityKeyBuf := buf.New()
		identityKey := identityKeyBuf.Extend(int32(method.GetSessionSubKeyAndSaltLength()))
		if mask == nil {
			err := derivation.GetIdentitySubKey(a.ipsk[current], salt, identityKey)
			if err != nil {
				return nil, newError("failed to get identity sub key").Base(err)
			}
		} else {
			copy(identityKey, a.ipsk[current])
		}
		eih[current] = [16]byte{}
		if mask != nil {
			subtle.XORBytes(currentPskHash[:], mask, currentPskHash[:])
		}
		err := method.GenerateEIH(identityKey, currentPskHash[:], eih[current][:])
		if err != nil {
			return nil, newError("failed to generate EIH").Base(err)
		}
		current--
		if current < 0 {
			break
		}
		currentPskHash = a.ipskHash[current]
		identityKeyBuf.Release()
	}
	return newAESEIHWithData(a.length, eih), nil
}
