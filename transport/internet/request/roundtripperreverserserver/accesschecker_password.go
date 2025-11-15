package roundtripperreverserserver

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"
	"math"

	"github.com/v2fly/struc"
)

type PasswordAccessChecker struct {
	Password string

	block cipher.Block
}

// NewPasswordAccessChecker creates a PasswordAccessChecker from the given password
// and initializes its internal cipher block. It returns an error if initialization
// fails.
func NewPasswordAccessChecker(password string) (*PasswordAccessChecker, error) {
	p := &PasswordAccessChecker{Password: password}
	if err := p.init(); err != nil {
		return nil, err
	}
	return p, nil
}

type tokenUnpacked struct {
	UserID int64  `struc:"int64,little"`
	Check  uint64 `struc:"uint64,little"`
}

func (p *PasswordAccessChecker) CheckReverserAccess(ctx context.Context, serverKey []byte) (clientKey []byte, err error) {
	if len(serverKey) != 16 {
		return nil, newError("invalid server key length")
	}
	buffer := make([]byte, 16)
	// Decrypt into buffer from serverKey
	p.block.Decrypt(buffer, serverKey)

	token := &tokenUnpacked{}

	err = struc.Unpack(bytes.NewReader(buffer), token)
	if err != nil {
		return nil, newError("failed to unpack token").Base(err)
	}

	expectedCheck := uint64(0)
	if token.Check != expectedCheck {
		return nil, newError("invalid token check value")
	}

	if token.UserID == int64(math.MinInt64) {
		return nil, newError("invalid token userID: MinInt64")
	}

	if token.UserID < 0 {
		return nil, newError("invalid token userID for server: client token has negative userID")
	}

	// pack client token into a buffer with capacity 16
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	err = struc.Pack(buf, &tokenUnpacked{
		UserID: -token.UserID,
		Check:  expectedCheck,
	})
	if err != nil {
		return nil, newError("failed to pack client token").Base(err)
	}
	clientKeyBuffer := buf.Bytes()
	if len(clientKeyBuffer) != 16 {
		return nil, newError("invalid packed client token length")
	}
	// encrypt into a fresh slice to avoid overlapping issues
	encrypted := make([]byte, 16)
	p.block.Encrypt(encrypted, clientKeyBuffer)
	return encrypted, nil
}

func (p *PasswordAccessChecker) GenerateToken(userID int64) ([]byte, error) {
	if userID == int64(math.MinInt64) {
		return nil, newError("userID cannot be MinInt64")
	}
	expectedCheck := uint64(0)
	// pack token into a buffer with capacity 16
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	err := struc.Pack(buf, &tokenUnpacked{
		UserID: userID,
		Check:  expectedCheck,
	})
	if err != nil {
		return nil, newError("failed to pack token").Base(err)
	}
	buffer := buf.Bytes()
	if len(buffer) != 16 {
		return nil, newError("invalid packed token length")
	}
	encrypted := make([]byte, 16)
	p.block.Encrypt(encrypted, buffer)
	return encrypted, nil
}

func (p *PasswordAccessChecker) init() error {
	block, err := createBlockFromPassword(p.Password)
	if err != nil {
		return newError("failed to create block from password").Base(err)
	}
	p.block = block
	return nil
}

func createBlockFromPassword(password string) (cipher.Block, error) {
	// derive a 16-byte key from the password with hkdf
	key, err := hkdf.Expand(sha256.New, []byte(password), "v2ray-hd87BYQL-aBzumdEh-Yv4E6Rdu:request-roundtripperreverserserver"+"createBlockFromPassword", 16)
	if err != nil {
		return nil, newError("unable to derive key from password").Base(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, newError("unable to create AES cipher").Base(err)
	}
	return block, nil
}
