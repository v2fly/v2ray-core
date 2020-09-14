package trojan_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	. "v2ray.com/core/proxy/trojan"
)

func toAccount(a *Account) protocol.Account {
	account, err := a.AsAccount()
	common.Must(err)
	return account
}

func TestTCPRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Email: "love@v2ray.com",
		Account: toAccount(&Account{
			Password: "password",
		}),
	}
	payload := []byte("test string")
	data := buf.New()
	common.Must2(data.Write(payload))

	buffer := buf.New()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_TCP, Address: net.LocalHostIP, Port: 1234}
	writer, err := WriteHeader(buffer, destination, user.Account.(*MemoryAccount))
	common.Must(err)
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	dest, err := ReadHeader(buffer)
	common.Must(err)

	if r := cmp.Diff(*dest, destination); r != "" {
		t.Error("destination: ", r)
	}

	reader := buf.NewReader(buffer)
	decodedData, err := reader.ReadMultiBuffer()
	common.Must(err)
	if r := cmp.Diff(decodedData[0].Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}

}

func TestUDPRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Email: "love@v2ray.com",
		Account: toAccount(&Account{
			Password: "password",
		}),
	}
	payload := []byte("test string")
	data := buf.New()
	common.Must2(data.Write(payload))

	buffer := buf.New()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_UDP, Address: net.LocalHostIP, Port: 1234}
	writer, err := WriteHeader(buffer, destination, user.Account.(*MemoryAccount))
	common.Must(err)
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	_, err = ReadHeader(buffer)
	common.Must(err)

	dest, mb, err := ReadPacket(buffer)
	common.Must(err)

	if dest == nil {
		t.Error("destination is empty")
	}

	if mb.IsEmpty() {
		t.Error("no request data")
	}

	if r := cmp.Diff(*dest, destination); r != "" {
		t.Error("destination: ", r)
	}

	mb, decoded := buf.SplitFirst(mb)
	buf.ReleaseMulti(mb)

	if r := cmp.Diff(decoded.Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}

}
