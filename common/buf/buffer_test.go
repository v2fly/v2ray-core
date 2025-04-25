package buf_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghxhy/v2ray-core/v5/common"
	. "github.com/ghxhy/v2ray-core/v5/common/buf"
)

func TestBufferClear(t *testing.T) {
	buffer := New()
	defer buffer.Release()

	payload := "Bytes"
	buffer.Write([]byte(payload))
	if diff := cmp.Diff(buffer.Bytes(), []byte(payload)); diff != "" {
		t.Error(diff)
	}

	buffer.Clear()
	if buffer.Len() != 0 {
		t.Error("expect 0 length, but got ", buffer.Len())
	}
}

func TestBufferIsEmpty(t *testing.T) {
	buffer := New()
	defer buffer.Release()

	if buffer.IsEmpty() != true {
		t.Error("expect empty buffer, but not")
	}
}

func TestBufferString(t *testing.T) {
	buffer := New()
	defer buffer.Release()

	const payload = "Test String"
	common.Must2(buffer.WriteString(payload))
	if buffer.String() != payload {
		t.Error("expect buffer content as ", payload, " but actually ", buffer.String())
	}
}

func TestBufferByte(t *testing.T) {
	{
		buffer := New()
		common.Must(buffer.WriteByte('m'))
		if buffer.String() != "m" {
			t.Error("expect buffer content as ", "m", " but actually ", buffer.String())
		}
		buffer.Release()
	}
	{
		buffer := StackNew()
		common.Must(buffer.WriteByte('n'))
		if buffer.String() != "n" {
			t.Error("expect buffer content as ", "n", " but actually ", buffer.String())
		}
		buffer.Release()
	}
	{
		buffer := StackNew()
		common.Must2(buffer.WriteString("HELLOWORLD"))
		if b := buffer.Byte(5); b != 'W' {
			t.Error("unexpected byte ", b)
		}

		buffer.SetByte(5, 'M')
		if buffer.String() != "HELLOMORLD" {
			t.Error("expect buffer content as ", "n", " but actually ", buffer.String())
		}
		buffer.Release()
	}
}

func TestBufferResize(t *testing.T) {
	buffer := New()
	defer buffer.Release()

	const payload = "Test String"
	common.Must2(buffer.WriteString(payload))
	if buffer.String() != payload {
		t.Error("expect buffer content as ", payload, " but actually ", buffer.String())
	}

	buffer.Resize(-6, -3)
	if l := buffer.Len(); int(l) != 3 {
		t.Error("len error ", l)
	}

	if s := buffer.String(); s != "Str" {
		t.Error("unexpect buffer ", s)
	}

	buffer.Resize(int32(len(payload)), 200)
	if l := buffer.Len(); int(l) != 200-len(payload) {
		t.Error("len error ", l)
	}
}

func TestBufferSlice(t *testing.T) {
	{
		b := New()
		common.Must2(b.Write([]byte("abcd")))
		bytes := b.BytesFrom(-2)
		if diff := cmp.Diff(bytes, []byte{'c', 'd'}); diff != "" {
			t.Error(diff)
		}
	}

	{
		b := New()
		common.Must2(b.Write([]byte("abcd")))
		bytes := b.BytesTo(-2)
		if diff := cmp.Diff(bytes, []byte{'a', 'b'}); diff != "" {
			t.Error(diff)
		}
	}

	{
		b := New()
		common.Must2(b.Write([]byte("abcd")))
		bytes := b.BytesRange(-3, -1)
		if diff := cmp.Diff(bytes, []byte{'b', 'c'}); diff != "" {
			t.Error(diff)
		}
	}
}

func TestBufferReadFullFrom(t *testing.T) {
	payload := make([]byte, 1024)
	common.Must2(rand.Read(payload))

	reader := bytes.NewReader(payload)
	b := New()
	n, err := b.ReadFullFrom(reader, 1024)
	common.Must(err)
	if n != 1024 {
		t.Error("expect reading 1024 bytes, but actually ", n)
	}

	if diff := cmp.Diff(payload, b.Bytes()); diff != "" {
		t.Error(diff)
	}
}

func BenchmarkNewBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := New()
		buffer.Release()
	}
}

func BenchmarkNewBufferStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := StackNew()
		buffer.Release()
	}
}

func BenchmarkWrite2(b *testing.B) {
	buffer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buffer.Write([]byte{'a', 'b'})
		buffer.Clear()
	}
}

func BenchmarkWrite8(b *testing.B) {
	buffer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buffer.Write([]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'})
		buffer.Clear()
	}
}

func BenchmarkWrite32(b *testing.B) {
	buffer := New()
	payload := make([]byte, 32)
	rand.Read(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buffer.Write(payload)
		buffer.Clear()
	}
}

func BenchmarkWriteByte2(b *testing.B) {
	buffer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buffer.WriteByte('a')
		_ = buffer.WriteByte('b')
		buffer.Clear()
	}
}

func BenchmarkWriteByte8(b *testing.B) {
	buffer := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buffer.WriteByte('a')
		_ = buffer.WriteByte('b')
		_ = buffer.WriteByte('c')
		_ = buffer.WriteByte('d')
		_ = buffer.WriteByte('e')
		_ = buffer.WriteByte('f')
		_ = buffer.WriteByte('g')
		_ = buffer.WriteByte('h')
		buffer.Clear()
	}
}
