package trojan_test

import (
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	. "github.com/v2fly/v2ray-core/v5/proxy/trojan"
)

func toAccount(a *Account) protocol.Account {
	account, err := a.AsAccount()
	common.Must(err)
	return account
}

func TestTCPRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Email: "love@v2fly.org",
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
	writer := &ConnWriter{Writer: buffer, Target: destination, Account: user.Account.(*MemoryAccount)}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	reader := &ConnReader{Reader: buffer}
	common.Must(reader.ParseHeader())

	if r := cmp.Diff(reader.Target, destination); r != "" {
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
		Email: "love@v2fly.org",
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
	writer := &PacketWriter{Writer: &ConnWriter{Writer: buffer, Target: destination, Account: user.Account.(*MemoryAccount)}, Target: destination}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data}))

	connReader := &ConnReader{Reader: buffer}
	common.Must(connReader.ParseHeader())

	packetReader := &PacketReader{Reader: connReader}
	p, err := packetReader.ReadMultiBufferWithMetadata()
	common.Must(err)

	if p.Buffer.IsEmpty() {
		t.Error("no request data")
	}

	if r := cmp.Diff(p.Target, destination); r != "" {
		t.Error("destination: ", r)
	}

	mb, decoded := buf.SplitFirst(p.Buffer)
	buf.ReleaseMulti(mb)

	if r := cmp.Diff(decoded.Bytes(), payload); r != "" {
		t.Error("data: ", r)
	}
}

func TestLargeUDPRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Email: "love@v2fly.org",
		Account: toAccount(&Account{
			Password: "password",
		}),
	}

	payload := make([]byte, 4096)
	common.Must2(rand.Read(payload))
	data := buf.NewWithSize(int32(len(payload)))
	common.Must2(data.Write(payload))

	buffer := buf.NewWithSize(2*data.Len() + 1)
	defer buffer.Release()

	destination := net.Destination{Network: net.Network_UDP, Address: net.LocalHostIP, Port: 1234}
	writer := &PacketWriter{Writer: &ConnWriter{Writer: buffer, Target: destination, Account: user.Account.(*MemoryAccount)}, Target: destination}
	common.Must(writer.WriteMultiBuffer(buf.MultiBuffer{data, data}))

	connReader := &ConnReader{Reader: buffer}
	common.Must(connReader.ParseHeader())

	packetReader := &PacketReader{Reader: connReader}
	for i := 0; i < 2; i++ {
		p, err := packetReader.ReadMultiBufferWithMetadata()
		common.Must(err)

		if p.Buffer.IsEmpty() {
			t.Error("no request data")
		}

		if r := cmp.Diff(p.Target, destination); r != "" {
			t.Error("destination: ", r)
		}

		mb, decoded := buf.SplitFirst(p.Buffer)
		buf.ReleaseMulti(mb)

		if r := cmp.Diff(decoded.Bytes(), payload); r != "" {
			t.Error("data: ", r)
		}
	}
}
