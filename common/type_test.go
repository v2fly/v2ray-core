package common_test

import (
	"context"
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/common"
)

type TConfig struct {
	value int
}

type YConfig struct {
	value string
}

func TestObjectCreation(t *testing.T) {
	f := func(ctx context.Context, t interface{}) (interface{}, error) {
		return func() int {
			return t.(*TConfig).value
		}, nil
	}

	Must(RegisterConfig((*TConfig)(nil), f))
	err := RegisterConfig((*TConfig)(nil), f)
	if err == nil {
		t.Error("expect non-nil error, but got nil")
	}

	g, err := CreateObject(context.Background(), &TConfig{value: 2})
	Must(err)
	if v := g.(func() int)(); v != 2 {
		t.Error("expect return value 2, but got ", v)
	}

	_, err = CreateObject(context.Background(), &YConfig{value: "T"})
	if err == nil {
		t.Error("expect non-nil error, but got nil")
	}
}
