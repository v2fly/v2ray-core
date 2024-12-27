package shadowsocks2022

import (
	"encoding/hex"
	"io"

	"github.com/v2fly/struc"
)

func newRequestSaltWithLength(length int) RequestSalt {
	return &requestSaltWithLength{length: length}
}

type requestSaltWithLength struct {
	length  int
	content []byte
}

func (r *requestSaltWithLength) isRequestSalt() {}
func (r *requestSaltWithLength) Pack(p []byte, opt *struc.Options) (int, error) {
	n := copy(p, r.content)
	if n != r.length {
		return 0, newError("failed to pack request salt with length")
	}
	return n, nil
}

func (r *requestSaltWithLength) Unpack(reader io.Reader, length int, opt *struc.Options) error {
	r.content = make([]byte, r.length)
	n, err := io.ReadFull(reader, r.content)
	if err != nil {
		return newError("failed to unpack request salt with length").Base(err)
	}
	if n != r.length {
		return newError("failed to unpack request salt with length")
	}
	return nil
}

func (r *requestSaltWithLength) Size(opt *struc.Options) int {
	return r.length
}

func (r *requestSaltWithLength) String() string {
	return hex.Dump(r.content)
}

func (r *requestSaltWithLength) Bytes() []byte {
	return r.content
}

func (r *requestSaltWithLength) FillAllFrom(reader io.Reader) error {
	r.content = make([]byte, r.length)
	_, err := io.ReadFull(reader, r.content)
	if err != nil {
		return newError("failed to fill salt from reader").Base(err)
	}
	return nil
}
