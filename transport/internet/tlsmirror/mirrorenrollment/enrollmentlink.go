package mirrorenrollment

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

const (
	// MIME type used for enrollment data URLs.
	enrollmentDataMIME = "application/vnd.v2ray.tlsmirror-enrollment"
)

// LinkFromAny converts a protobuf Any into a data URL string.
// The produced link has form:
// data:application/vnd.v2ray.tlsmirror-enrollment;base64,<base64(payload)>
// where payload is the marshaled Any message encoded with standard base64 (with padding).
func LinkFromAny(a *anypb.Any) (string, error) {
	// Machine generated code
	if a == nil {
		return "", newError("nil Any")
	}
	b, err := proto.Marshal(a)
	if err != nil {
		return "", newError("failed to marshal Any").Base(err)
	}
	enc := base64.StdEncoding.EncodeToString(b)
	dataURL := "data:" + enrollmentDataMIME + ";base64," + enc
	return dataURL, nil
}

// AnyFromLink converts a string link (now primarily a data URL) back to *anypb.Any.
// Accepted formats (strict): only data URLs matching the exact MIME type and base64 encoding.
func AnyFromLink(link string) (*anypb.Any, error) {
	// Machine generated code
	if link == "" {
		return nil, newError("empty link")
	}

	// Must be a data URL.
	if !strings.HasPrefix(link, "data:") {
		return nil, newError("input must be a data URL")
	}

	// Parse and split header and payload at the first comma.
	comma := strings.Index(link, ",")
	if comma == -1 || comma+1 >= len(link) {
		return nil, newError("invalid data URL: missing payload")
	}

	meta := link[len("data:"):comma]
	payload := link[comma+1:]

	// Meta should be like "<mime-type>;base64" possibly with additional params.
	metaParts := strings.Split(meta, ";")
	if len(metaParts) < 2 {
		return nil, newError("invalid data URL metadata")
	}

	// First part must match our expected MIME.
	if metaParts[0] != enrollmentDataMIME {
		return nil, newError("unexpected MIME type in data URL: " + metaParts[0])
	}

	// Ensure "base64" is present in parameters.
	isBase64 := false
	for _, p := range metaParts[1:] {
		if p == "base64" {
			isBase64 = true
			break
		}
	}
	if !isBase64 {
		return nil, newError("data URL must be base64 encoded")
	}

	// No URL parsing/decoding of payload: payload is raw base64.
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, newError("failed to base64-decode data URL payload").Base(err)
	}

	var any anypb.Any
	if err := proto.Unmarshal(raw, &any); err != nil {
		return nil, newError("failed to unmarshal Any from decoded payload").Base(err)
	}
	return &any, nil
}

// helper to format an error when newError is not available at call site
func fmtErr(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}
