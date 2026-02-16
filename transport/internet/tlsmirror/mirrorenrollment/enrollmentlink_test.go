package mirrorenrollment

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func TestLinkRoundTrip(t *testing.T) {
	orig := &anypb.Any{
		TypeUrl: "type.test/example",
		Value:   []byte("sample-payload-bytes"),
	}

	link, err := LinkFromAny(orig)
	if err != nil {
		t.Fatalf("LinkFromAny failed: %v", err)
	}
	if !strings.HasPrefix(link, "data:") {
		t.Fatalf("expected data URL prefix, got: %s", link)
	}

	decoded, err := AnyFromLink(link)
	if err != nil {
		t.Fatalf("AnyFromLink failed: %v", err)
	}
	if !proto.Equal(decoded, orig) {
		t.Fatalf("decoded Any does not match original.\nOriginal: %#v\nDecoded: %#v", orig, decoded)
	}
}

func TestLinkFromAnyProducesValidDataURL(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/inspect", Value: []byte("inspect-bytes")}
	link, err := LinkFromAny(orig)
	if err != nil {
		t.Fatalf("LinkFromAny failed: %v", err)
	}
	if !strings.HasPrefix(link, "data:") {
		t.Fatalf("expected data URL prefix, got: %s", link)
	}
	comma := strings.Index(link, ",")
	if comma == -1 {
		t.Fatalf("no comma in data URL: %s", link)
	}
	meta := link[len("data:"):comma]
	payload := link[comma+1:]
	if !strings.HasPrefix(meta, enrollmentDataMIME) {
		t.Fatalf("unexpected MIME in meta: %s", meta)
	}
	if !strings.Contains(meta, "base64") {
		t.Fatalf("expected base64 param in meta: %s", meta)
	}

	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("failed to decode payload with std encoding: %v", err)
	}
	var any anypb.Any
	if err := proto.Unmarshal(raw, &any); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if !proto.Equal(&any, orig) {
		t.Fatalf("unmarshaled Any does not equal original\nwant=%v\ngot=%v", orig, &any)
	}
}

func TestLegacyTLSMirrorScheme(t *testing.T) {
	orig := &anypb.Any{
		TypeUrl: "type.test/legacy",
		Value:   []byte("legacy-bytes"),
	}
	marshaled, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("failed to marshal orig Any: %v", err)
	}
	enc := base64.RawURLEncoding.EncodeToString(marshaled)
	legacyLink := "tlsmirror-enrollment:" + enc

	if _, err := AnyFromLink(legacyLink); err == nil {
		t.Fatalf("expected AnyFromLink to reject legacy link format, but it succeeded")
	}
}

func TestPlainBase64Input(t *testing.T) {
	orig := &anypb.Any{
		TypeUrl: "type.test/plain",
		Value:   []byte("plain-bytes"),
	}
	marshaled, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("failed to marshal orig Any: %v", err)
	}
	enc := base64.StdEncoding.EncodeToString(marshaled)
	// Pass plain base64 (no scheme) - should be rejected in strict mode
	if _, err := AnyFromLink(enc); err == nil {
		t.Fatalf("expected AnyFromLink to reject plain base64 input, but it succeeded")
	}
}

func TestInvalidInput(t *testing.T) {
	if _, err := AnyFromLink("not-a-valid-base64@@"); err == nil {
		t.Fatalf("expected error for invalid input, got nil")
	}

	if _, err := LinkFromAny(nil); err == nil {
		t.Fatalf("expected error for LinkFromAny(nil), got nil")
	}
}

func TestRejectWrongMIME(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/wrongmime", Value: []byte("x")}
	b, _ := proto.Marshal(orig)
	enc := base64.StdEncoding.EncodeToString(b)
	link := fmt.Sprintf("data:application/octet-stream;base64,%s", enc)
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected rejection for wrong MIME type, but got success")
	}
}

func TestRejectMissingBase64Param(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/nobase64", Value: []byte("x")}
	b, _ := proto.Marshal(orig)
	enc := base64.StdEncoding.EncodeToString(b)
	link := fmt.Sprintf("data:%s,%s", enrollmentDataMIME, enc) // no ;base64
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected rejection for missing base64 param, but got success")
	}
}

func TestAcceptExtraParams(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/params", Value: []byte("p")}
	b, _ := proto.Marshal(orig)
	enc := base64.StdEncoding.EncodeToString(b)
	link := fmt.Sprintf("data:%s;v=1;base64;foo=bar,%s", enrollmentDataMIME, enc)
	any, err := AnyFromLink(link)
	if err != nil {
		t.Fatalf("expected success for extra params, got error: %v", err)
	}
	if !proto.Equal(any, orig) {
		t.Fatalf("decoded Any mismatch; want %v got %v", orig, any)
	}
}

func TestRejectNonBase64Payload(t *testing.T) {
	link := "data:" + enrollmentDataMIME + ";base64,not-base64-@@@"
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected error for non-base64 payload, got nil")
	}
}

func TestRejectURLEncodedPayload(t *testing.T) {
	// percent-encoded payload
	orig := &anypb.Any{TypeUrl: "type.test/urlenc", Value: []byte("url-bytes")}
	b, _ := proto.Marshal(orig)
	enc := base64.StdEncoding.EncodeToString(b)
	// percent-encode first few chars
	pct := "%" + enc[:2] + enc[2:]
	link := fmt.Sprintf("data:%s;base64,%s", enrollmentDataMIME, pct)
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected rejection for percent-encoded payload, got success")
	}
}

func TestCaseSensitiveMIME(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/case", Value: []byte("c")}
	b, _ := proto.Marshal(orig)
	enc := base64.StdEncoding.EncodeToString(b)
	link := fmt.Sprintf("data:%s;base64,%s", strings.ToUpper(enrollmentDataMIME), enc)
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected rejection for case mismatch MIME, got success")
	}
}

func TestLargePayloadRoundTrip(t *testing.T) {
	// create a ~100KB payload
	size := 100 * 1024
	p := make([]byte, size)
	if _, err := rand.Read(p); err != nil {
		t.Fatalf("failed to generate random payload: %v", err)
	}
	orig := &anypb.Any{TypeUrl: "type.test/large", Value: p}
	link, err := LinkFromAny(orig)
	if err != nil {
		t.Fatalf("LinkFromAny failed: %v", err)
	}
	any, err := AnyFromLink(link)
	if err != nil {
		t.Fatalf("AnyFromLink failed for large payload: %v", err)
	}
	if !proto.Equal(any, orig) {
		t.Fatalf("large payload mismatch after roundtrip")
	}
}

func TestRejectRawURLEncodingInDataURL(t *testing.T) {
	orig := &anypb.Any{TypeUrl: "type.test/rawurl", Value: []byte("raw")}
	b, _ := proto.Marshal(orig)
	enc := base64.RawURLEncoding.EncodeToString(b)
	link := fmt.Sprintf("data:%s;base64,%s", enrollmentDataMIME, enc)
	if _, err := AnyFromLink(link); err == nil {
		t.Fatalf("expected rejection for raw URL base64 payload, got success")
	}
}
