package shadowsocksr_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	. "github.com/v2fly/v2ray-core/v5/proxy/shadowsocksr"
)

func toAccount(a *Account) protocol.Account {
	account, err := a.AsAccount()
	common.Must(err)
	return account
}

func equalRequestHeader(x, y *protocol.RequestHeader) bool {
	return cmp.Equal(x, y, cmp.Comparer(func(x, y protocol.RequestHeader) bool {
		return x == y
	}))
}

func TestTCPRequest(t *testing.T) {
	cases := []struct {
		name    string
		request *protocol.RequestHeader
		payload []byte
	}{
		{
			name: "AES-256-CFB with auth_aes128_md5",
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.LocalHostIP,
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:      "tcp-password",
						CipherType:    CipherType_AES_256_CFB,
						Protocol:      "auth_aes128_md5",
						ProtocolParam: "64",
						Obfs:         "tls1.2_ticket_auth",
						ObfsParam:    "cloudflare.com",
					}),
				},
			},
			payload: []byte("test string"),
		},
		{
			name: "CHACHA20 with auth_chain_a",
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.LocalHostIPv6,
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:      "password",
						CipherType:    CipherType_CHACHA20,
						Protocol:      "auth_chain_a",
						ProtocolParam: "",
						Obfs:         "http_simple",
						ObfsParam:    "microsoft.com",
					}),
				},
			},
			payload: []byte("test string"),
		},
		{
			name: "RC4-MD5 with origin protocol",
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.DomainAddress("v2fly.org"),
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:   "password",
						CipherType: CipherType_RC4_MD5,
						Protocol:  "origin",
						Obfs:      "plain",
					}),
				},
			},
			payload: []byte("test string"),
		},
	}

	runTest := func(request *protocol.RequestHeader, payload []byte) {
		data := buf.New()
		common.Must2(data.Write(payload))

		cache := buf.New()
		defer cache.Release()

		writer, err := WriteTCPRequest(request, cache)
		common.Must(err)

		common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

		decodedRequest, reader, err := ReadTCPSession(request.User, cache)
		common.Must(err)
		if equalRequestHeader(decodedRequest, request) == false {
			t.Error("different request")
		}

		decodedData, err := reader.ReadMultiBuffer()
		common.Must(err)
		if r := cmp.Diff(decodedData[0].Bytes(), payload); r != "" {
			t.Error("data: ", r)
		}
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			runTest(test.request, test.payload)
		})
	}
}

func TestTCPReaderWriter(t *testing.T) {
	cases := []struct {
		name    string
		account *Account
	}{
		{
			name: "AES-256-CFB with auth_aes128_md5",
			account: &Account{
				Password:      "test-password",
				CipherType:    CipherType_AES_256_CFB,
				Protocol:      "auth_aes128_md5",
				ProtocolParam: "64",
				Obfs:         "tls1.2_ticket_auth",
				ObfsParam:    "cloudflare.com",
			},
		},
		{
			name: "CHACHA20 with auth_chain_a",
			account: &Account{
				Password:      "test-password",
				CipherType:    CipherType_CHACHA20,
				Protocol:      "auth_chain_a",
				ProtocolParam: "",
				Obfs:         "http_simple",
				ObfsParam:    "microsoft.com",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			user := &protocol.MemoryUser{
				Account: toAccount(test.account),
			}
			cache := buf.New()
			defer cache.Release()

			writer, err := WriteTCPRequest(&protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.DomainAddress("v2fly.org"),
				Port:    123,
				User:    user,
			}, cache)
			common.Must(err)

			_, reader, err := ReadTCPSession(user, cache)
			common.Must(err)

			// Test multiple writes
			messages := []string{"test payload", "test payload 2"}
			for _, msg := range messages {
				b := buf.New()
				common.Must2(b.WriteString(msg))
				common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{b}))

				data, err := reader.ReadMultiBuffer()
				common.Must(err)
				if data[0].String() != msg {
					t.Error("unexpected output: ", data[0].String())
				}
			}
		})
	}
}
