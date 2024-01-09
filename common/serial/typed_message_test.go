package serial_test

import (
	"testing"

	. "github.com/v2fly/v2ray-core/v5/common/serial"
)

func TestGetInstance(t *testing.T) {
	p, err := GetInstance("")
	if p != nil {
		t.Error("expected nil instance, but got ", p)
	}
	if err == nil {
		t.Error("expect non-nil error, but got nil")
	}
}

func TestConvertingNilMessage(t *testing.T) {
	x := ToTypedMessage(nil)
	if x != nil {
		t.Error("expect nil, but actually not")
	}
}

func TestGetMessageDescriptor(t *testing.T) {
	_, err := GetMessageDescriptor("")
	if err == nil {
		t.Error("expect non-nil error, but got nil")
	}

	md, err := GetMessageDescriptor("google.protobuf.Any")
	if err != nil {
		t.Error("expect nil error, but got ", err)
	}

	if md == nil || md.FullName() != "google.protobuf.Any" {
		t.Error("expect google.protobuf.Any, but got ", md)
	}
}
