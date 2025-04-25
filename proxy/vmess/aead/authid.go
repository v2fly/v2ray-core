package aead

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	rand3 "crypto/rand"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"math"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/antireplay"
)

var (
	ErrNotFound = errors.New("user do not exist")
	ErrReplay   = errors.New("replayed request")
)

func CreateAuthID(cmdKey []byte, time int64) [16]byte {
	buf := bytes.NewBuffer(nil)
	common.Must(binary.Write(buf, binary.BigEndian, time))
	var zero uint32
	common.Must2(io.CopyN(buf, rand3.Reader, 4))
	zero = crc32.ChecksumIEEE(buf.Bytes())
	common.Must(binary.Write(buf, binary.BigEndian, zero))
	aesBlock := NewCipherFromKey(cmdKey)
	if buf.Len() != 16 {
		panic("Size unexpected")
	}
	var result [16]byte
	aesBlock.Encrypt(result[:], buf.Bytes())
	return result
}

func NewCipherFromKey(cmdKey []byte) cipher.Block {
	aesBlock, err := aes.NewCipher(KDF16(cmdKey, KDFSaltConstAuthIDEncryptionKey))
	if err != nil {
		panic(err)
	}
	return aesBlock
}

type AuthIDDecoder struct {
	s cipher.Block
}

func NewAuthIDDecoder(cmdKey []byte) *AuthIDDecoder {
	return &AuthIDDecoder{NewCipherFromKey(cmdKey)}
}

func (aidd *AuthIDDecoder) Decode(data [16]byte) (int64, uint32, int32, []byte) {
	aidd.s.Decrypt(data[:], data[:])
	var t int64
	var zero uint32
	var rand int32
	reader := bytes.NewReader(data[:])
	common.Must(binary.Read(reader, binary.BigEndian, &t))
	common.Must(binary.Read(reader, binary.BigEndian, &rand))
	common.Must(binary.Read(reader, binary.BigEndian, &zero))
	return t, zero, rand, data[:]
}

func NewAuthIDDecoderHolder() *AuthIDDecoderHolder {
	return &AuthIDDecoderHolder{make(map[string]*AuthIDDecoderItem), antireplay.NewReplayFilter(120)}
}

type AuthIDDecoderHolder struct {
	decoders map[string]*AuthIDDecoderItem
	filter   *antireplay.ReplayFilter
}

type AuthIDDecoderItem struct {
	dec    *AuthIDDecoder
	ticket interface{}
}

func NewAuthIDDecoderItem(key [16]byte, ticket interface{}) *AuthIDDecoderItem {
	return &AuthIDDecoderItem{
		dec:    NewAuthIDDecoder(key[:]),
		ticket: ticket,
	}
}

func (a *AuthIDDecoderHolder) AddUser(key [16]byte, ticket interface{}) {
	a.decoders[string(key[:])] = NewAuthIDDecoderItem(key, ticket)
}

func (a *AuthIDDecoderHolder) RemoveUser(key [16]byte) {
	delete(a.decoders, string(key[:]))
}

func (a *AuthIDDecoderHolder) Match(authID [16]byte) (interface{}, error) {
	for _, v := range a.decoders {
		t, z, _, d := v.dec.Decode(authID)
		if z != crc32.ChecksumIEEE(d[:12]) {
			continue
		}

		if t < 0 {
			continue
		}

		if math.Abs(math.Abs(float64(t))-float64(time.Now().Unix())) > 120 {
			continue
		}

		if !a.filter.Check(authID[:]) {
			return nil, ErrReplay
		}

		return v.ticket, nil
	}
	return nil, ErrNotFound
}
