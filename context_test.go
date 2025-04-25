package core_test

import (
	"context"
	"testing"
	_ "unsafe"

	. "github.com/ghxhy/v2ray-core/v5"
)

func TestFromContextPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expect panic, but nil")
		}
	}()

	MustFromContext(context.Background())
}
