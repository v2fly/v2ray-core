package http

import (
	"bufio"
	"io"
	"net/http"
	"strings"
	"testing"
)

// A malformed upstream response whose first line is shorter than 4 bytes used to
// make readResponseAndHandle100Continue slice ResponseHeader1xx[len-4:] with a
// negative lower bound, panicking the whole process.
func TestReadResponseAndHandle100ContinueShortLine(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	// "\n 1" is parsed as a 1xx status line, but the first ReadSlice('\n')
	// only yields a single byte.
	payload := "\n 1\r\n" + strings.Repeat("A", 100)
	reader := bufio.NewReader(strings.NewReader(payload))

	// The response is malformed, so an error is expected; a panic is not.
	_, err = readResponseAndHandle100Continue(reader, req, io.Discard)
	if err == nil {
		t.Error("expected an error for a malformed response")
	}
}

func TestReadResponseAndHandle100Continue(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	payload := "HTTP/1.1 100 Continue\r\n\r\nHTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(payload))

	forwarded := &strings.Builder{}
	resp, err := readResponseAndHandle100Continue(reader, req, forwarded)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("expected the 200 response, got ", resp.StatusCode)
	}
	if forwarded.String() != "HTTP/1.1 100 Continue\r\n\r\n" {
		t.Error("expected the 1xx response to be forwarded, got ", forwarded.String())
	}
}
