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

// errUnwrapSentinel is a sentinel for exercising errors.Is traversal.
var errUnwrapSentinel = stderrors.New("unwrap sentinel")

// typedUnwrapErr is a concrete error type for exercising errors.As traversal.
type typedUnwrapErr struct{ code int }

func (typedUnwrapErr) Error() string { return "typed unwrap error" }

func TestErrorUnwrap(t *testing.T) {
	base := stderrors.New("base")
	if got := stderrors.Unwrap(New("wrapper").Base(base)); got != base {
		t.Errorf("Unwrap() = %v, want the wrapped base error", got)
	}
	if got := stderrors.Unwrap(New("no inner")); got != nil {
		t.Errorf("Unwrap() of an error with no inner = %v, want nil", got)
	}
}

func TestErrorIsThroughChain(t *testing.T) {
	// Three *Error layers, matching the real-world depth of a wrapped error.
	chain := New("a").Base(New("b").Base(New("c").Base(errUnwrapSentinel)))
	if !stderrors.Is(chain, errUnwrapSentinel) {
		t.Errorf("errors.Is could not find the sentinel through the *Error chain: %v", chain)
	}
}

func TestErrorAsThroughChain(t *testing.T) {
	chain := New("outer").Base(New("inner").Base(typedUnwrapErr{code: 42}))
	var got typedUnwrapErr
	if !stderrors.As(chain, &got) {
		t.Fatalf("errors.As could not extract typedUnwrapErr through the *Error chain: %v", chain)
	}
	if got.code != 42 {
		t.Errorf("extracted typedUnwrapErr.code = %d, want 42", got.code)
	}
}

// TestErrorAsSyscallErrno mirrors the motivating case: a syscall.Errno from a
// failed bind, wrapped by three *Error layers (as ListenTCP -> ListenTCP ->
// worker.Start do), must remain recoverable via the standard errors.As.
func TestErrorAsSyscallErrno(t *testing.T) {
	const want = syscall.Errno(4242) // arbitrary; only round-trip traversal is asserted
	chain := New("failed to listen TCP on 443").Base(
		New("failed to listen on address: 0.0.0.0:443").Base(
			New("failed to listen TCP on 0.0.0.0:443").Base(want)))
	var got syscall.Errno
	if !stderrors.As(chain, &got) {
		t.Fatalf("errors.As could not extract syscall.Errno through the *Error chain: %v", chain)
	}
	if got != want {
		t.Errorf("extracted errno = %d, want %d", got, want)
	}
}
