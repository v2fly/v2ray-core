package protocol_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	. "github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/uuid"
	"github.com/ghxhy/v2ray-core/v5/proxy/vmess"
)

func TestAlwaysValidStrategy(t *testing.T) {
	strategy := AlwaysValid()
	if !strategy.IsValid() {
		t.Error("strategy not valid")
	}
	strategy.Invalidate()
	if !strategy.IsValid() {
		t.Error("strategy not valid")
	}
}

func TestTimeoutValidStrategy(t *testing.T) {
	strategy := BeforeTime(time.Now().Add(2 * time.Second))
	if !strategy.IsValid() {
		t.Error("strategy not valid")
	}
	time.Sleep(3 * time.Second)
	if strategy.IsValid() {
		t.Error("strategy is valid")
	}

	strategy = BeforeTime(time.Now().Add(2 * time.Second))
	strategy.Invalidate()
	if strategy.IsValid() {
		t.Error("strategy is valid")
	}
}

func TestUserInServerSpec(t *testing.T) {
	uuid1 := uuid.New()
	uuid2 := uuid.New()

	toAccount := func(a *vmess.Account) Account {
		account, err := a.AsAccount()
		common.Must(err)
		return account
	}

	spec := NewServerSpec(net.Destination{}, AlwaysValid(), &MemoryUser{
		Email:   "test1@v2fly.org",
		Account: toAccount(&vmess.Account{Id: uuid1.String()}),
	})
	if spec.HasUser(&MemoryUser{
		Email:   "test1@v2fly.org",
		Account: toAccount(&vmess.Account{Id: uuid2.String()}),
	}) {
		t.Error("has user: ", uuid2)
	}

	spec.AddUser(&MemoryUser{Email: "test2@v2fly.org"})
	if !spec.HasUser(&MemoryUser{
		Email:   "test1@v2fly.org",
		Account: toAccount(&vmess.Account{Id: uuid1.String()}),
	}) {
		t.Error("not having user: ", uuid1)
	}
}

func TestPickUser(t *testing.T) {
	spec := NewServerSpec(net.Destination{}, AlwaysValid(), &MemoryUser{Email: "test1@v2fly.org"}, &MemoryUser{Email: "test2@v2fly.org"}, &MemoryUser{Email: "test3@v2fly.org"})
	user := spec.PickUser()
	if !strings.HasSuffix(user.Email, "@v2fly.org") {
		t.Error("user: ", user.Email)
	}
}
