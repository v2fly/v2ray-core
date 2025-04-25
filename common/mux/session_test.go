package mux_test

import (
	"testing"

	. "github.com/ghxhy/v2ray-core/v5/common/mux"
)

func TestSessionManagerAdd(t *testing.T) {
	m := NewSessionManager()

	s := m.Allocate()
	if s.ID != 1 {
		t.Error("id: ", s.ID)
	}
	if m.Size() != 1 {
		t.Error("size: ", m.Size())
	}

	s = m.Allocate()
	if s.ID != 2 {
		t.Error("id: ", s.ID)
	}
	if m.Size() != 2 {
		t.Error("size: ", m.Size())
	}

	s = &Session{
		ID: 4,
	}
	m.Add(s)
	if s.ID != 4 {
		t.Error("id: ", s.ID)
	}
	if m.Size() != 3 {
		t.Error("size: ", m.Size())
	}
}

func TestSessionManagerClose(t *testing.T) {
	m := NewSessionManager()
	s := m.Allocate()

	if m.CloseIfNoSession() {
		t.Error("able to close")
	}
	m.Remove(s.ID)
	if !m.CloseIfNoSession() {
		t.Error("not able to close")
	}
}
