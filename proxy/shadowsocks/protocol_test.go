package shadowsocks_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	. "github.com/ghxhy/v2ray-core/v5/proxy/shadowsocks"
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

func TestUDPEncoding(t *testing.T) {
	request := &protocol.RequestHeader{
		Version: Version,
		Command: protocol.RequestCommandUDP,
		Address: net.LocalHostIP,
		Port:    1234,
		User: &protocol.MemoryUser{
			Email: "love@v2fly.org",
			Account: toAccount(&Account{
				Password:   "password",
				CipherType: CipherType_AES_128_GCM,
			}),
		},
	}

	data := buf.New()
	common.Must2(data.WriteString("test string"))
	encodedData, err := EncodeUDPPacket(request, data.Bytes())
	common.Must(err)

	decodedRequest, decodedData, err := DecodeUDPPacket(request.User, encodedData)
	common.Must(err)

	if r := cmp.Diff(decodedData.Bytes(), data.Bytes()); r != "" {
		t.Error("data: ", r)
	}

	if equalRequestHeader(decodedRequest, request) == false {
		t.Error("different request")
	}
}

func TestTCPRequest(t *testing.T) {
	cases := []struct {
		request *protocol.RequestHeader
		payload []byte
	}{
		{
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.LocalHostIP,
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:   "tcp-password",
						CipherType: CipherType_AES_128_GCM,
					}),
				},
			},
			payload: []byte("test string"),
		},
		{
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.LocalHostIPv6,
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:   "password",
						CipherType: CipherType_AES_256_GCM,
					}),
				},
			},
			payload: []byte("test string"),
		},
		{
			request: &protocol.RequestHeader{
				Version: Version,
				Command: protocol.RequestCommandTCP,
				Address: net.DomainAddress("v2fly.org"),
				Port:    1234,
				User: &protocol.MemoryUser{
					Email: "love@v2fly.org",
					Account: toAccount(&Account{
						Password:   "password",
						CipherType: CipherType_CHACHA20_POLY1305,
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
		runTest(test.request, test.payload)
	}
}

func TestUDPReaderWriter(t *testing.T) {
	user := &protocol.MemoryUser{
		Account: toAccount(&Account{
			Password:   "test-password",
			CipherType: CipherType_CHACHA20_POLY1305,
		}),
	}
	cache := buf.New()
	defer cache.Release()

	writer := &buf.SequentialWriter{Writer: &UDPWriter{
		Writer: cache,
		Request: &protocol.RequestHeader{
			Version: Version,
			Address: net.DomainAddress("v2fly.org"),
			Port:    123,
			User:    user,
		},
	}}

	reader := &UDPReader{
		Reader: cache,
		User:   user,
	}

	{
		b := buf.New()
		common.Must2(b.WriteString("test payload"))
		common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{b}))

		payload, err := reader.ReadMultiBuffer()
		common.Must(err)
		if payload[0].String() != "test payload" {
			t.Error("unexpected output: ", payload[0].String())
		}
	}

	{
		b := buf.New()
		common.Must2(b.WriteString("test payload 2"))
		common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{b}))

		payload, err := reader.ReadMultiBuffer()
		common.Must(err)
		if payload[0].String() != "test payload 2" {
			t.Error("unexpected output: ", payload[0].String())
		}
	}
}
