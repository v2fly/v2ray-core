package shadowsocksr_test

import (
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/proxy/shadowsocksr"
)

func TestStreamCipherUDP(t *testing.T) {
	testCases := []struct {
		name       string
		cipherType shadowsocksr.CipherType
	}{
		{
			name:       "AES-256-CFB",
			cipherType: shadowsocksr.CipherType_AES_256_CFB,
		},
		{
			name:       "AES-128-CFB",
			cipherType: shadowsocksr.CipherType_AES_128_CFB,
		},
		{
			name:       "CHACHA20",
			cipherType: shadowsocksr.CipherType_CHACHA20,
		},
		{
			name:       "RC4-MD5",
			cipherType: shadowsocksr.CipherType_RC4_MD5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawAccount := &shadowsocksr.Account{
				Password:   "test",
				CipherType: tc.cipherType,
				Protocol:  "origin",
				Obfs:      "plain",
			}
			account, err := rawAccount.AsAccount()
			common.Must(err)

			cipher := account.(*shadowsocksr.MemoryAccount).Cipher

			key := make([]byte, cipher.KeySize())
			common.Must2(rand.Read(key))

			payload := make([]byte, 1024)
			common.Must2(rand.Read(payload))

			b1 := buf.New()
			common.Must2(b1.ReadFullFrom(rand.Reader, cipher.IVSize()))
			common.Must2(b1.Write(payload))
			common.Must(cipher.EncodePacket(key, b1))

			common.Must(cipher.DecodePacket(key, b1))
			if diff := cmp.Diff(b1.Bytes(), payload); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestProtocolEncryption(t *testing.T) {
	testCases := []struct {
		name       string
		cipherType shadowsocksr.CipherType
		protocol   string
		protocolParam string
	}{
		{
			name:       "AES-256-CFB with auth_aes128_md5",
			cipherType: shadowsocksr.CipherType_AES_256_CFB,
			protocol:   "auth_aes128_md5",
			protocolParam: "64",
		},
		{
			name:       "CHACHA20 with auth_chain_a",
			cipherType: shadowsocksr.CipherType_CHACHA20,
			protocol:   "auth_chain_a",
			protocolParam: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawAccount := &shadowsocksr.Account{
				Password:      "test",
				CipherType:    tc.cipherType,
				Protocol:     tc.protocol,
				ProtocolParam: tc.protocolParam,
				Obfs:         "plain",
			}
			account, err := rawAccount.AsAccount()
			common.Must(err)

			cipher := account.(*shadowsocksr.MemoryAccount).Cipher

			key := make([]byte, cipher.KeySize())
			common.Must2(rand.Read(key))

			payload := make([]byte, 1024)
			common.Must2(rand.Read(payload))

			b1 := buf.New()
			common.Must2(b1.ReadFullFrom(rand.Reader, cipher.IVSize()))
			common.Must2(b1.Write(payload))
			common.Must(cipher.EncodePacket(key, b1))

			common.Must(cipher.DecodePacket(key, b1))
			if diff := cmp.Diff(b1.Bytes(), payload); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestObfsEncryption(t *testing.T) {
	testCases := []struct {
		name       string
		cipherType shadowsocksr.CipherType
		obfs       string
		obfsParam  string
	}{
		{
			name:       "AES-256-CFB with tls1.2_ticket_auth",
			cipherType: shadowsocksr.CipherType_AES_256_CFB,
			obfs:       "tls1.2_ticket_auth",
			obfsParam:  "cloudflare.com",
		},
		{
			name:       "CHACHA20 with http_simple",
			cipherType: shadowsocksr.CipherType_CHACHA20,
			obfs:       "http_simple",
			obfsParam:  "microsoft.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawAccount := &shadowsocksr.Account{
				Password:   "test",
				CipherType: tc.cipherType,
				Protocol:  "origin",
				Obfs:      tc.obfs,
				ObfsParam: tc.obfsParam,
			}
			account, err := rawAccount.AsAccount()
			common.Must(err)

			cipher := account.(*shadowsocksr.MemoryAccount).Cipher

			key := make([]byte, cipher.KeySize())
			common.Must2(rand.Read(key))

			payload := make([]byte, 1024)
			common.Must2(rand.Read(payload))

			b1 := buf.New()
			common.Must2(b1.ReadFullFrom(rand.Reader, cipher.IVSize()))
			common.Must2(b1.Write(payload))
			common.Must(cipher.EncodePacket(key, b1))

			common.Must(cipher.DecodePacket(key, b1))
			if diff := cmp.Diff(b1.Bytes(), payload); diff != "" {
				t.Error(diff)
			}
		})
	}
}
