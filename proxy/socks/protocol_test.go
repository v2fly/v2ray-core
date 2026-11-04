package socks_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	. "github.com/v2fly/v2ray-core/v5/proxy/socks"
)

func TestClientHandshakeRequestAddress(t *testing.T) {
	testCases := []struct {
		name    string
		address net.Address
		encoded string
	}{
		{
			name:    "IPv4",
			address: net.ParseAddress("198.51.100.9"),
			encoded: "01c633640901bb",
		},
		{
			name:    "IPv6",
			address: net.ParseAddress("2001:db8::9"),
			encoded: "0420010db800000000000000000000000901bb",
		},
		{
			name:    "domain",
			address: net.DomainAddress("example.com"),
			encoded: "030b6578616d706c652e636f6d01bb",
		},
	}

	for _, command := range []protocol.RequestCommand{protocol.RequestCommandTCP, protocol.RequestCommandUDP} {
		name := "TCP"
		if command == protocol.RequestCommandUDP {
			name = "UDP"
		}
		for _, testCase := range testCases {
			t.Run(name+"/"+testCase.name, func(t *testing.T) {
				request := &protocol.RequestHeader{
					Version: 0x05,
					Command: command,
					Address: testCase.address,
					Port:    443,
				}
				// The server selects no authentication and reports bound endpoint 192.0.2.1:1235.
				reader := bytes.NewReader([]byte{0x05, 0x00, 0x05, 0x00, 0x00, 0x01, 192, 0, 2, 1, 0x04, 0xd3})
				var writer bytes.Buffer
				relay, err := ClientHandshake(request, reader, &writer, false)
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := hex.DecodeString(testCase.encoded)
				if err != nil {
					t.Fatal(err)
				}
				want := append([]byte{0x05, 0x01, 0x00, 0x05, 0x01, 0x00}, encoded...)
				if command == protocol.RequestCommandUDP {
					// RFC 1928: an unknown UDP client endpoint must be all zeros.
					want = []byte{0x05, 0x01, 0x00, 0x05, 0x03, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
					if relay == nil || relay.Command != protocol.RequestCommandUDP ||
						!relay.Address.IP().Equal(net.ParseAddress("192.0.2.1").IP()) || relay.Port != 1235 {
						t.Fatalf("unexpected UDP relay: %v", relay)
					}
				} else if relay != nil {
					t.Fatalf("TCP handshake returned a UDP relay: %v", relay)
				}
				if !bytes.Equal(writer.Bytes(), want) {
					t.Errorf("handshake bytes = %x, want %x", writer.Bytes(), want)
				}
				if request.Address != testCase.address || request.Port != 443 || request.Command != command {
					t.Fatal("handshake modified the request destination")
				}
			})
		}
	}
}

func TestUDPEncoding(t *testing.T) {
	b := buf.New()

	request := &protocol.RequestHeader{
		Address: net.IPAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}),
		Port:    1024,
	}
	writer := &buf.SequentialWriter{Writer: NewUDPWriter(request, b)}

	content := []byte{'a'}
	payload := buf.New()
	payload.Write(content)
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{payload}))

	reader := NewUDPReader(b)

	decodedPayload, err := reader.ReadMultiBuffer()
	common.Must(err)
	if r := cmp.Diff(decodedPayload[0].Bytes(), content); r != "" {
		t.Error(r)
	}
}

func TestReadUsernamePassword(t *testing.T) {
	testCases := []struct {
		Input    []byte
		Username string
		Password string
		Error    bool
	}{
		{
			Input:    []byte{0x05, 0x01, 'a', 0x02, 'b', 'c'},
			Username: "a",
			Password: "bc",
		},
		{
			Input: []byte{0x05, 0x18, 'a', 0x02, 'b', 'c'},
			Error: true,
		},
	}

	for _, testCase := range testCases {
		reader := bytes.NewReader(testCase.Input)
		username, password, err := ReadUsernamePassword(reader)
		if testCase.Error {
			if err == nil {
				t.Error("for input: ", testCase.Input, " expect error, but actually nil")
			}
		} else {
			if err != nil {
				t.Error("for input: ", testCase.Input, " expect no error, but actually ", err.Error())
			}
			if testCase.Username != username {
				t.Error("for input: ", testCase.Input, " expect username ", testCase.Username, " but actually ", username)
			}
			if testCase.Password != password {
				t.Error("for input: ", testCase.Input, " expect password ", testCase.Password, " but actually ", password)
			}
		}
	}
}

func TestReadUntilNull(t *testing.T) {
	testCases := []struct {
		Input  []byte
		Output string
		Error  bool
	}{
		{
			Input:  []byte{'a', 'b', 0x00},
			Output: "ab",
		},
		{
			Input: []byte{'a'},
			Error: true,
		},
	}

	for _, testCase := range testCases {
		reader := bytes.NewReader(testCase.Input)
		value, err := ReadUntilNull(reader)
		if testCase.Error {
			if err == nil {
				t.Error("for input: ", testCase.Input, " expect error, but actually nil")
			}
		} else {
			if err != nil {
				t.Error("for input: ", testCase.Input, " expect no error, but actually ", err.Error())
			}
			if testCase.Output != value {
				t.Error("for input: ", testCase.Input, " expect output ", testCase.Output, " but actually ", value)
			}
		}
	}
}

func BenchmarkReadUsernamePassword(b *testing.B) {
	input := []byte{0x05, 0x01, 'a', 0x02, 'b', 'c'}
	buffer := buf.New()
	buffer.Write(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ReadUsernamePassword(buffer)
		common.Must(err)
		buffer.Clear()
		buffer.Extend(int32(len(input)))
	}
}
