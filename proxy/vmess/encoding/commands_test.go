package encoding_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"v2ray.com/core/v4/common"
	"v2ray.com/core/v4/common/buf"
	"v2ray.com/core/v4/common/protocol"
	"v2ray.com/core/v4/common/uuid"
	. "v2ray.com/core/v4/proxy/vmess/encoding"
)

func TestSwitchAccount(t *testing.T) {
	sa := &protocol.CommandSwitchAccount{
		Port:     1234,
		ID:       uuid.New(),
		AlterIds: 1024,
		Level:    128,
		ValidMin: 16,
	}

	buffer := buf.New()
	common.Must(MarshalCommand(sa, buffer))

	cmd, err := UnmarshalCommand(1, buffer.BytesFrom(2))
	common.Must(err)

	sa2, ok := cmd.(*protocol.CommandSwitchAccount)
	if !ok {
		t.Fatal("failed to convert command to CommandSwitchAccount")
	}
	if r := cmp.Diff(sa2, sa); r != "" {
		t.Error(r)
	}
}
