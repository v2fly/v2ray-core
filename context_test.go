package core_test

import (
	"context"
	"testing"

	. "v2ray.com/core/v4"
)

func TestContextPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expect panic, but nil")
		}
	}()

	MustFromContext(context.Background())
}
