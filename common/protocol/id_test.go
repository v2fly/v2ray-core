package protocol_test

import (
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/uuid"
)

func TestIdEquals(t *testing.T) {
	id1 := NewID(uuid.New())
	id2 := NewID(id1.UUID())

	if !id1.Equals(id2) {
		t.Error("expected id1 to equal id2, but actually not")
	}

	if id1.String() != id2.String() {
		t.Error(id1.String(), " != ", id2.String())
	}
}
