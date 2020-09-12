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

	buffer := buf.StackNew()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_TCP, Address: net.LocalHostIP, Port: 1234}
	writer, err := WriteHeader(&buffer, destination, user.Account.(*MemoryAccount))
	common.Must(err)
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	dest, reader, err := ReadHeader(&buffer)
	common.Must(err)

	if r := cmp.Diff(*dest, destination); r != "" {
		t.Error("destination: ", r)
	}

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

	buffer := buf.StackNew()
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_UDP, Address: net.LocalHostIP, Port: 1234}
	writer, err := WriteHeader(&buffer, destination, user.Account.(*MemoryAccount))
	common.Must(err)
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	dest, reader, err := ReadHeader(&buffer)
	common.Must(err)

	if r := cmp.Diff(*dest, destination); r != "" {
		t.Error("destination: ", r)
	}

	decodedData, err := reader.ReadMultiBuffer()
	common.Must(err)
	if r := cmp.Diff(decodedData[0].Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}

}
