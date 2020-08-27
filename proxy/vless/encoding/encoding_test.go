package encoding_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/v2fly/v2ray-core/common"
	"github.com/v2fly/v2ray-core/common/buf"
	"github.com/v2fly/v2ray-core/common/net"
	"github.com/v2fly/v2ray-core/common/protocol"
	"github.com/v2fly/v2ray-core/common/uuid"
	"github.com/v2fly/v2ray-core/proxy/vless"
	. "github.com/v2fly/v2ray-core/proxy/vless/encoding"
)

func toAccount(a *vless.Account) protocol.Account {
	account, err := a.AsAccount()
	common.Must(err)
	return account
}

func TestRequestSerialization(t *testing.T) {
	user := &protocol.MemoryUser{
		Level: 0,
		Email: "test@v2fly.org",
	}
	id := uuid.New()
	account := &vless.Account{
		Id: id.String(),
	}
	user.Account = toAccount(account)

	expectedRequest := &protocol.RequestHeader{
		Version: Version,
		User:    user,
		Command: protocol.RequestCommandTCP,
		Address: net.DomainAddress("www.v2fly.org"),
		Port:    net.Port(443),
	}
	expectedAddons := &Addons{}

	buffer := buf.StackNew()
	common.Must(EncodeRequestHeader(&buffer, expectedRequest, expectedAddons))

	Validator := new(vless.Validator)
	Validator.Add(user)

	actualRequest, actualAddons, err, _ := DecodeRequestHeader(&buffer, Validator)
	common.Must(err)

	if r := cmp.Diff(actualRequest, expectedRequest, cmp.AllowUnexported(protocol.ID{})); r != "" {
		t.Error(r)
	}
	if r := cmp.Diff(actualAddons, expectedAddons); r != "" {
		t.Error(r)
	}
}

func TestInvalidRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Level: 0,
		Email: "test@v2fly.org",
	}
	id := uuid.New()
	account := &vless.Account{
		Id: id.String(),
	}
	user.Account = toAccount(account)

	expectedRequest := &protocol.RequestHeader{
		Version: Version,
		User:    user,
		Command: protocol.RequestCommand(100),
		Address: net.DomainAddress("www.v2fly.org"),
		Port:    net.Port(443),
	}
	expectedAddons := &Addons{}

	buffer := buf.StackNew()
	common.Must(EncodeRequestHeader(&buffer, expectedRequest, expectedAddons))

	Validator := new(vless.Validator)
	Validator.Add(user)

	_, _, err, _ := DecodeRequestHeader(&buffer, Validator)
	if err == nil {
		t.Error("nil error")
	}
}

func TestMuxRequest(t *testing.T) {
	user := &protocol.MemoryUser{
		Level: 0,
		Email: "test@v2fly.org",
	}
	id := uuid.New()
	account := &vless.Account{
		Id: id.String(),
	}
	user.Account = toAccount(account)

	expectedRequest := &protocol.RequestHeader{
		Version: Version,
		User:    user,
		Command: protocol.RequestCommandMux,
		Address: net.DomainAddress("v1.mux.cool"),
	}
	expectedAddons := &Addons{}

	buffer := buf.StackNew()
	common.Must(EncodeRequestHeader(&buffer, expectedRequest, expectedAddons))

	Validator := new(vless.Validator)
	Validator.Add(user)

	actualRequest, actualAddons, err, _ := DecodeRequestHeader(&buffer, Validator)
	common.Must(err)

	if r := cmp.Diff(actualRequest, expectedRequest, cmp.AllowUnexported(protocol.ID{})); r != "" {
		t.Error(r)
	}
	if r := cmp.Diff(actualAddons, expectedAddons); r != "" {
		t.Error(r)
	}
}
