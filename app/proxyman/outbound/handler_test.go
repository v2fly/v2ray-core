package outbound_test

import (
	"testing"

	. "v2ray.com/core/v4/app/proxyman/outbound"
	"v2ray.com/core/v4/features/outbound"
)

func TestInterfaces(t *testing.T) {
	_ = (outbound.Handler)(new(Handler))
	_ = (outbound.Manager)(new(Manager))
}
