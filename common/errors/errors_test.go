package errors_test

import (
	stderrors "errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"

	. "github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/log"
)

func TestError(t *testing.T) {
	err := New("TestError")
	if v := GetSeverity(err); v != log.Severity_Info {
		t.Error("severity: ", v)
	}

	err = New("TestError2").Base(io.EOF)
	if v := GetSeverity(err); v != log.Severity_Info {
		t.Error("severity: ", v)
	}

	err = New("TestError3").Base(io.EOF).AtWarning()
	if v := GetSeverity(err); v != log.Severity_Warning {
		t.Error("severity: ", v)
	}

	err = New("TestError4").Base(io.EOF).AtWarning()
	err = New("TestError5").Base(err)
	if v := GetSeverity(err); v != log.Severity_Warning {
		t.Error("severity: ", v)
	}
	if v := err.Error(); !strings.Contains(v, "EOF") {
		t.Error("error: ", v)
	}
}

type e struct{}

func TestErrorMessage(t *testing.T) {
	data := []struct {
		err error
		msg string
	}{
		{
			err: New("a").Base(New("b")).WithPathObj(e{}),
			msg: "common/errors_test: a > b",
		},
		{
			err: New("a").Base(New("b").WithPathObj(e{})),
			msg: "a > common/errors_test: b",
		},
	}

	for _, d := range data {
		if diff := cmp.Diff(d.msg, d.err.Error()); diff != "" {
			t.Error(diff)
		}
	}
}

var errUnwrapSentinel = stderrors.New("unwrap sentinel")

type typedUnwrapErr struct{ code int }

func (typedUnwrapErr) Error() string { return "typed unwrap error" }

func TestErrorUnwrap(t *testing.T) {
	expected := stderrors.New("base")
	if actual := stderrors.Unwrap(New("wrapper").Base(expected)); actual != expected {
		t.Errorf("Unwrap() = %v, expected the wrapped base error", actual)
	}
	if actual := stderrors.Unwrap(New("no inner")); actual != nil {
		t.Errorf("Unwrap() = %v, expected nil", actual)
	}
}

func TestErrorIsThroughChain(t *testing.T) {
	chain := New("a").Base(New("b").Base(New("c").Base(errUnwrapSentinel)))
	if !stderrors.Is(chain, errUnwrapSentinel) {
		t.Errorf("errors.Is could not find the sentinel through the *Error chain: %v", chain)
	}
}

func TestErrorAsThroughChain(t *testing.T) {
	chain := New("outer").Base(New("inner").Base(typedUnwrapErr{code: 42}))
	var actual typedUnwrapErr
	if !stderrors.As(chain, &actual) {
		t.Fatalf("errors.As could not extract typedUnwrapErr through the *Error chain: %v", chain)
	}
	if actual.code != 42 {
		t.Errorf("extracted code = %d, expected 42", actual.code)
	}
}

// The motivating case: recover a syscall.Errno from a multi-layer wrap.
func TestErrorAsSyscallErrno(t *testing.T) {
	const expected = syscall.Errno(4242)
	chain := New("failed to listen TCP on 443").Base(
		New("failed to listen on address: 0.0.0.0:443").Base(
			New("failed to listen TCP on 0.0.0.0:443").Base(expected)))
	var actual syscall.Errno
	if !stderrors.As(chain, &actual) {
		t.Fatalf("errors.As could not extract syscall.Errno through the *Error chain: %v", chain)
	}
	if actual != expected {
		t.Errorf("extracted errno = %d, expected %d", actual, expected)
	}
}
