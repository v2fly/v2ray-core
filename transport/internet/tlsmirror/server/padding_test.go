package server

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestPackUnpackRoundTrip(t *testing.T) {
	payload := []byte("v2ray")
	data, padding := Unpack(Pack(append([]byte(nil), payload...), 7))
	if !bytes.Equal(data, payload) {
		t.Fatalf("got payload %q, want %q", data, payload)
	}
	if padding != 7 {
		t.Fatalf("got padding %v, want 7", padding)
	}
}

// TestUnpackRejectsOversizedLength makes sure that a hostile length field is
// rejected instead of being converted into a negative slice bound, which
// panics on platforms where int is 32 bits wide.
func TestUnpackRejectsOversizedLength(t *testing.T) {
	for _, length := range []uint32{0xffffffff, 0x80000000, 0x7fffffff, 16} {
		wrapped := make([]byte, 16)
		binary.BigEndian.PutUint32(wrapped[len(wrapped)-4:], length)
		data, padding := Unpack(wrapped)
		if data != nil {
			t.Fatalf("length %v returned %v bytes of data, want none", length, len(data))
		}
		if padding != 0 {
			t.Fatalf("length %v returned padding %v, want 0", length, padding)
		}
	}
}
