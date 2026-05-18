package gdocsviewer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func TestBuildOriginRequestURLPlaintext(t *testing.T) {
	c := &client{config: &ClientConfig{
		OriginUrl: "https://{abc}-origin.test/gdocsviewer/",
		OriginUrlReplacementRules: []*OriginUrlReplacementRule{{
			Name:    "abc",
			Pattern: "[a-z0-9]{14}",
		}},
	}}

	got, err := c.buildOriginRequestURL(request.Request{
		ConnectionTag: []byte{1, 2, 3},
		Data:          []byte("payload"),
	})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := neturl.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(parsed.Host, "-origin.test") {
		t.Fatalf("unexpected generated host %q", parsed.Host)
	}
	generated := strings.TrimSuffix(parsed.Host, "-origin.test")
	if len(generated) != 14 || !allBytesIn(generated, "abcdefghijklmnopqrstuvwxyz0123456789") {
		t.Fatalf("unexpected generated label %q", generated)
	}
	parts := strings.Split(strings.TrimPrefix(parsed.Path, "/gdocsviewer/r/"), "/")
	if len(parts) != 3 {
		t.Fatalf("unexpected path %q", parsed.Path)
	}
	if parts[0] != "AQID" {
		t.Fatalf("unexpected session segment %q", parts[0])
	}
	if parts[1] != "cGF5bG9hZA" {
		t.Fatalf("unexpected payload segment %q", parts[1])
	}
	if !strings.HasSuffix(parts[2], ".txt") || strings.TrimSuffix(parts[2], ".txt") == "" {
		t.Fatalf("unexpected nonce segment %q", parts[2])
	}
}

func TestBuildOriginRequestURLEncrypted(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	c := &client{config: &ClientConfig{
		OriginUrl: "https://origin-{abc}.test/base",
		SharedKey: key,
		OriginUrlReplacementRules: []*OriginUrlReplacementRule{{
			Name:    "abc",
			Pattern: "[a-f0-9]{8}",
		}},
	}}

	got, err := c.buildOriginRequestURL(request.Request{
		ConnectionTag: []byte("session"),
		Data:          []byte("payload"),
	})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := neturl.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(parsed.Host, "origin-") || !strings.HasSuffix(parsed.Host, ".test") {
		t.Fatalf("unexpected generated host %q", parsed.Host)
	}
	encoded := strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/base/t/"), ".log")
	combined, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	session, payload, err := decryptClientRequest(key, combined)
	if err != nil {
		t.Fatal(err)
	}
	if string(session) != "session" || string(payload) != "payload" {
		t.Fatalf("unexpected decrypted request: session=%q payload=%q", session, payload)
	}
}

func TestOriginURLReplacementUsesOneValuePerRule(t *testing.T) {
	c := &client{config: &ClientConfig{
		OriginUrl: "https://{abc}.{abc}.origin.test/g",
		OriginUrlReplacementRules: []*OriginUrlReplacementRule{{
			Name:    "abc",
			Pattern: "[a-z0-9]{14}",
		}},
	}}

	got, err := c.buildOriginRequestURL(request.Request{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := neturl.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	labels := strings.Split(parsed.Host, ".")
	if len(labels) < 2 || labels[0] != labels[1] {
		t.Fatalf("expected repeated placeholder to share one generated value, host=%q", parsed.Host)
	}
}

func TestOriginURLReplacementAllowsValidUnusedRule(t *testing.T) {
	c := &client{config: &ClientConfig{
		OriginUrl: "https://origin.test/g",
		OriginUrlReplacementRules: []*OriginUrlReplacementRule{{
			Name:    "unused",
			Pattern: "[a-z0-9]{14}",
		}},
	}}

	got, err := c.buildOriginRequestURL(request.Request{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := neturl.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "origin.test" {
		t.Fatalf("unexpected host %q", parsed.Host)
	}
}

func TestOriginURLReplacementRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name string
		rule *OriginUrlReplacementRule
	}{
		{name: "invalid name", rule: &OriginUrlReplacementRule{Name: "bad/name", Pattern: "[a-z]{4}"}},
		{name: "invalid pattern", rule: &OriginUrlReplacementRule{Name: "abc", Pattern: "[a-z"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &client{config: &ClientConfig{
				OriginUrl:                 "https://origin.test/g",
				OriginUrlReplacementRules: []*OriginUrlReplacementRule{test.rule},
			}}
			if _, err := c.buildOriginRequestURL(request.Request{}); err == nil {
				t.Fatal("expected invalid replacement rule error")
			}
		})
	}
}

func TestGenerateOriginURLReplacementPatternSyntax(t *testing.T) {
	got, err := generateOriginURLReplacement(&OriginUrlReplacementRule{Name: "abc", Pattern: "[a-z0-9]{14}"}, zeroReader{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 14 || !allBytesIn(got, "abcdefghijklmnopqrstuvwxyz0123456789") {
		t.Fatalf("unexpected generated value %q", got)
	}

	got, err = generateOriginURLReplacement(&OriginUrlReplacementRule{Name: "abc", Pattern: `edge\-\{x\}`}, zeroReader{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "edge-{x}" {
		t.Fatalf("unexpected escaped literal output %q", got)
	}

	got, err = generateOriginURLReplacement(&OriginUrlReplacementRule{Name: "abc", Pattern: "edge-[a-f0-9]{8}"}, zeroReader{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "edge-") || len(got) != len("edge-")+8 || !allBytesIn(strings.TrimPrefix(got, "edge-"), "abcdef0123456789") {
		t.Fatalf("unexpected generated value %q", got)
	}
}

func TestGenerateOriginURLReplacementRejectsUnsupportedPatterns(t *testing.T) {
	tests := []string{
		"",
		"[a-z",
		"[z-a]",
		"[a-z]{0}",
		"[a-z]{257}",
		"(abc)",
		"a|b",
		"[a-z]+",
		"[a-z]{1,2}",
	}
	for _, pattern := range tests {
		t.Run(pattern, func(t *testing.T) {
			_, err := generateOriginURLReplacement(&OriginUrlReplacementRule{Name: "abc", Pattern: pattern}, zeroReader{})
			if err == nil {
				t.Fatal("expected pattern error")
			}
		})
	}
}

func TestRoundTripPlaintextViewerFlowAndHeaders(t *testing.T) {
	rt := &recordingRoundTripper{}
	rt.fn = func(req *http.Request) (*http.Response, error) {
		switch len(rt.requests) {
		case 1:
			if req.URL.Host != "viewer.test" || req.URL.Path != "/viewer" {
				t.Fatalf("unexpected viewer request URL %s", req.URL.String())
			}
			origin := req.URL.Query().Get("url")
			originURL, err := neturl.Parse(origin)
			if err != nil {
				t.Fatal(err)
			}
			parts := strings.Split(strings.TrimPrefix(originURL.Path, "/gdocsviewer/r/"), "/")
			if len(parts) != 3 || parts[0] != "c2Vzc2lvbg" || parts[1] != "cmVxdWVzdA" || !strings.HasSuffix(parts[2], ".txt") {
				t.Fatalf("unexpected origin request URL %q", origin)
			}
			return textHTTPResponse(`... "/viewerng/text?id=doc-1&authuser=0" ...`), nil
		case 2:
			if req.URL.Path != "/viewerng/text" || req.URL.Query().Get("id") != "doc-1" || req.URL.Query().Get("page") != "0" {
				t.Fatalf("unexpected text request URL %s", req.URL.String())
			}
			return textHTTPResponse(viewerTextBodyForServerBody([]byte("response"))), nil
		default:
			t.Fatalf("unexpected request count %d", len(rt.requests))
			return nil, nil
		}
	}
	c := &client{httpRTT: rt, config: &ClientConfig{
		ViewerUrl:        "https://viewer.test/viewer",
		TextUrl:          "https://viewer.test/viewerng/text",
		OriginUrl:        "https://origin.test/gdocsviewer",
		ViewerHostHeader: "docs.example",
		UserAgent:        "gdocsviewer-test",
		RequestHeaders: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
			"Host":            "docs.custom",
			"User-Agent":      "gdocsviewer-custom-agent",
			"X-Gdocs-Test":    "1",
		},
		MinRequestIntervalMs: 1,
	}}

	resp, err := c.RoundTrip(context.Background(), request.Request{
		ConnectionTag: []byte("session"),
		Data:          []byte("request"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Data) != "response" {
		t.Fatalf("unexpected response %q", resp.Data)
	}
	if len(rt.requests) != 2 {
		t.Fatalf("unexpected request count %d", len(rt.requests))
	}
	for _, req := range rt.requests {
		if req.Host != "docs.custom" || req.Header.Get("User-Agent") != "gdocsviewer-custom-agent" {
			t.Fatalf("headers not applied: host=%q ua=%q", req.Host, req.Header.Get("User-Agent"))
		}
		if req.Header.Get("Accept-Language") != "en-US,en;q=0.9" || req.Header.Get("X-Gdocs-Test") != "1" {
			t.Fatalf("custom headers not applied: accept-language=%q x-gdocs-test=%q", req.Header.Get("Accept-Language"), req.Header.Get("X-Gdocs-Test"))
		}
	}
}

func TestExtractDocumentIDVariants(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "raw", body: `"/viewerng/text?id=raw-doc&x=1"`, want: "raw-doc"},
		{name: "relative text endpoint", body: `"text?id=relative-doc","status?id=relative-doc"`, want: "relative-doc"},
		{name: "url escaped", body: `viewerng%2Ftext%3Fid%3Durl-doc%26x%3D1`, want: "url-doc"},
		{name: "js slash escaped", body: `viewerng\/text?id=js-doc&x=1`, want: "js-doc"},
		{name: "js unicode escaped", body: `viewerng\u002Ftext\u003Fid\u003Dunicode-doc\u0026x\u003D1`, want: "unicode-doc"},
		{name: "js unicode relative", body: `text?id\u003dunicode-relative-doc`, want: "unicode-relative-doc"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractDocumentID([]byte(test.body))
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestDecodeViewerTextData(t *testing.T) {
	originBody := base64.StdEncoding.EncodeToString([]byte("response"))
	wrapped := viewerTextBodyForOriginBody(originBody)

	got, err := decodeViewerTextData([]byte(wrapped))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != originBody {
		t.Fatalf("got %q, want %q", got, originBody)
	}

	got, err = decodeViewerTextData([]byte(originBody))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != originBody {
		t.Fatalf("got %q, want %q", got, originBody)
	}
}

func TestRoundTripEncryptedErrorFrame(t *testing.T) {
	key := bytes.Repeat([]byte{9}, 32)
	encryptedError, err := encryptResponseFrame(key, append([]byte{responseFrameError}, []byte("blocked")...))
	if err != nil {
		t.Fatal(err)
	}
	rt := &recordingRoundTripper{}
	rt.fn = func(req *http.Request) (*http.Response, error) {
		if len(rt.requests) == 1 {
			return textHTTPResponse(`viewerng/text?id=doc-error`), nil
		}
		return textHTTPResponse(viewerTextBodyForOriginBody(base64.StdEncoding.EncodeToString(encryptedError))), nil
	}
	c := &client{httpRTT: rt, config: &ClientConfig{
		ViewerUrl:            "https://viewer.test/viewer",
		TextUrl:              "https://viewer.test/viewerng/text",
		OriginUrl:            "https://origin.test/gdocsviewer",
		SharedKey:            key,
		MinRequestIntervalMs: 1,
	}}

	_, err = c.RoundTrip(context.Background(), request.Request{Data: []byte("request")})
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("expected encrypted error frame, got %v", err)
	}
}

func TestRoundTripRejectsMalformedViewerData(t *testing.T) {
	rt := &recordingRoundTripper{}
	rt.fn = func(req *http.Request) (*http.Response, error) {
		if len(rt.requests) == 1 {
			return textHTTPResponse(`viewerng/text?id=doc-bad`), nil
		}
		return textHTTPResponse(`)]}'` + "\n" + `{"mimetype":"text/plain","data":`), nil
	}
	c := &client{httpRTT: rt, config: &ClientConfig{
		ViewerUrl:            "https://viewer.test/viewer",
		TextUrl:              "https://viewer.test/viewerng/text",
		OriginUrl:            "https://origin.test/gdocsviewer",
		MinRequestIntervalMs: 1,
	}}

	_, err := c.RoundTrip(context.Background(), request.Request{})
	if err == nil {
		t.Fatal("expected malformed viewer JSON error")
	}
}

func TestRoundTripRejectsInvalidOriginBase64Body(t *testing.T) {
	rt := &recordingRoundTripper{}
	rt.fn = func(req *http.Request) (*http.Response, error) {
		if len(rt.requests) == 1 {
			return textHTTPResponse(`viewerng/text?id=doc-invalid`), nil
		}
		return textHTTPResponse(viewerTextBodyForOriginBody("not base64")), nil
	}
	c := &client{httpRTT: rt, config: &ClientConfig{
		ViewerUrl:            "https://viewer.test/viewer",
		TextUrl:              "https://viewer.test/viewerng/text",
		OriginUrl:            "https://origin.test/gdocsviewer",
		MinRequestIntervalMs: 1,
	}}

	_, err := c.RoundTrip(context.Background(), request.Request{})
	if err == nil {
		t.Fatal("expected invalid origin base64 error")
	}
}

func TestServerPlaintextSuccessAndErrors(t *testing.T) {
	receiver := &fakeTripperReceiver{
		fn: func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
			return request.Response{Data: []byte("pong")}, nil
		},
	}
	s := &server{
		config:   &ServerConfig{PathPrefix: "/g", MaxRequestBytes: 10},
		assembly: fakeServerAssembly{receiver: receiver},
	}

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/g/r/"+base64.RawURLEncoding.EncodeToString([]byte("session"))+"/"+base64.RawURLEncoding.EncodeToString([]byte("ping"))+"/nonce.txt", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rec.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "pong" {
		t.Fatalf("unexpected response body %q", decoded)
	}
	if receiver.calls != 1 || string(receiver.last.ConnectionTag) != "session" || string(receiver.last.Data) != "ping" {
		t.Fatalf("unexpected receiver state calls=%d req=%+v", receiver.calls, receiver.last)
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/g/r/c2Vzc2lvbg/"+base64.RawURLEncoding.EncodeToString([]byte("too large payload"))+"/nonce.txt", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected request-size rejection 404, got %d", rec.Code)
	}

	receiver.fn = func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
		return request.Response{}, errors.New("backend failed")
	}
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/g/r/c2Vzc2lvbg/cGluZw/nonce.txt", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected backend error 404, got %d", rec.Code)
	}
}

func TestServerRoundTripIgnoresHTTPRequestContextCancellation(t *testing.T) {
	receiver := &fakeTripperReceiver{
		fn: func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
			if err := ctx.Err(); err != nil {
				return request.Response{}, fmt.Errorf("roundtrip context is canceled: %w", err)
			}
			return request.Response{Data: []byte("pong")}, nil
		},
	}
	s := &server{
		ctx:      context.Background(),
		config:   &ServerConfig{PathPrefix: "/g"},
		assembly: fakeServerAssembly{receiver: receiver},
	}

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	httpReq := httptest.NewRequest(http.MethodGet, "/g/r/c2Vzc2lvbg/cGluZw/nonce.txt", nil).WithContext(canceledContext)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httpReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rec.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "pong" {
		t.Fatalf("unexpected response body %q", decoded)
	}
}

func TestServerModeGating(t *testing.T) {
	key := bytes.Repeat([]byte{3}, 32)

	plain := &server{config: &ServerConfig{PathPrefix: "/g"}, assembly: fakeServerAssembly{receiver: &fakeTripperReceiver{}}}
	rec := httptest.NewRecorder()
	plain.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/g/t/not-encrypted.log", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("plaintext server accepted encrypted path: %d", rec.Code)
	}

	encrypted := &server{config: &ServerConfig{PathPrefix: "/g", SharedKey: key}, assembly: fakeServerAssembly{receiver: &fakeTripperReceiver{}}}
	rec = httptest.NewRecorder()
	encrypted.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/g/r/cw/cA/nonce.txt", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("encrypted server accepted plaintext path: %d", rec.Code)
	}
}

func TestServerEncryptedSuccessErrorAndLimits(t *testing.T) {
	key := bytes.Repeat([]byte{4}, 32)
	receiver := &fakeTripperReceiver{
		fn: func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
			return request.Response{Data: []byte("encrypted-pong")}, nil
		},
	}
	s := &server{
		config:   &ServerConfig{PathPrefix: "/g", SharedKey: key, MaxRequestBytes: 10, MaxResponseBytes: 20},
		assembly: fakeServerAssembly{receiver: receiver},
	}

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, encryptedServerRequest(t, key, []byte("session"), []byte("ping")))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	frame := decryptServerHTTPBody(t, key, rec.Body.String())
	if frame[0] != responseFrameSuccess || string(frame[1:]) != "encrypted-pong" {
		t.Fatalf("unexpected encrypted success frame %q", frame)
	}
	if receiver.calls != 1 || string(receiver.last.ConnectionTag) != "session" || string(receiver.last.Data) != "ping" {
		t.Fatalf("unexpected receiver state calls=%d req=%+v", receiver.calls, receiver.last)
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, encryptedServerRequest(t, key, []byte("session"), []byte("too large payload")))
	frame = decryptServerHTTPBody(t, key, rec.Body.String())
	if rec.Code != http.StatusOK || frame[0] != responseFrameError || !strings.Contains(string(frame[1:]), "max_request_bytes") {
		t.Fatalf("unexpected encrypted size error status=%d frame=%q", rec.Code, frame)
	}

	receiver.fn = func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
		return request.Response{}, errors.New("backend failed")
	}
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, encryptedServerRequest(t, key, []byte("session"), []byte("ping")))
	frame = decryptServerHTTPBody(t, key, rec.Body.String())
	if rec.Code != http.StatusOK || frame[0] != responseFrameError || !strings.Contains(string(frame[1:]), "backend failed") {
		t.Fatalf("unexpected encrypted backend error status=%d frame=%q", rec.Code, frame)
	}

	receiver.fn = func(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
		return request.Response{Data: []byte("this response is too large")}, nil
	}
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, encryptedServerRequest(t, key, []byte("session"), []byte("ping")))
	frame = decryptServerHTTPBody(t, key, rec.Body.String())
	if rec.Code != http.StatusOK || frame[0] != responseFrameError || !strings.Contains(string(frame[1:]), "max_response_bytes") {
		t.Fatalf("unexpected encrypted response-size error status=%d frame=%q", rec.Code, frame)
	}
}

func TestServerEncryptedAEADMismatchSignalsEncryptedError(t *testing.T) {
	serverKey := bytes.Repeat([]byte{5}, 32)
	clientKey := bytes.Repeat([]byte{6}, 32)
	s := &server{
		config:   &ServerConfig{PathPrefix: "/g", SharedKey: serverKey},
		assembly: fakeServerAssembly{receiver: &fakeTripperReceiver{}},
	}

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, encryptedServerRequest(t, clientKey, []byte("session"), []byte("ping")))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	frame := decryptServerHTTPBody(t, serverKey, rec.Body.String())
	if frame[0] != responseFrameError {
		t.Fatalf("unexpected frame %q", frame)
	}
	if _, err := decryptResponseFrame(clientKey, mustDecodeStdBase64(t, rec.Body.String())); err == nil {
		t.Fatal("expected client key mismatch to fail decryption")
	}
}

type recordingRoundTripper struct {
	requests []*http.Request
	fn       func(*http.Request) (*http.Response, error)
}

func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.requests = append(r.requests, req.Clone(req.Context()))
	return r.fn(req)
}

func textHTTPResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func viewerTextBodyForServerBody(body []byte) string {
	return viewerTextBodyForOriginBody(base64.StdEncoding.EncodeToString(body))
}

func viewerTextBodyForOriginBody(originBody string) string {
	body, err := json.Marshal(viewerTextResponse{Mimetype: "text/plain", Data: originBody})
	if err != nil {
		panic(err)
	}
	return ")]}'\n" + string(body)
}

type fakeServerAssembly struct {
	receiver request.TripperReceiver
}

func (f fakeServerAssembly) TripperReceiver() request.TripperReceiver {
	return f.receiver
}

func (f fakeServerAssembly) SessionReceiver() request.SessionReceiver {
	return nil
}

func (f fakeServerAssembly) AutoImplListener() request.Listener {
	return nil
}

type fakeTripperReceiver struct {
	calls int
	last  request.Request
	fn    func(context.Context, request.Request, ...request.RoundTripperOption) (request.Response, error)
}

func (f *fakeTripperReceiver) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (request.Response, error) {
	f.calls++
	f.last = request.Request{
		Data:          append([]byte(nil), req.Data...),
		ConnectionTag: append([]byte(nil), req.ConnectionTag...),
	}
	if f.fn != nil {
		return f.fn(ctx, req, opts...)
	}
	return request.Response{}, nil
}

func encryptedServerRequest(t *testing.T, key []byte, session []byte, payload []byte) *http.Request {
	t.Helper()
	combined, err := encryptClientRequest(key, session, payload)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewRequest(http.MethodGet, "/g/t/"+base64.RawURLEncoding.EncodeToString(combined)+".log", nil)
}

func decryptServerHTTPBody(t *testing.T, key []byte, body string) []byte {
	t.Helper()
	combined := mustDecodeStdBase64(t, body)
	frame, err := decryptResponseFrame(key, combined)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame) == 0 {
		t.Fatal("empty frame")
	}
	return frame
}

func mustDecodeStdBase64(t *testing.T, body string) []byte {
	t.Helper()
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body))
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

type zeroReader struct{}

func (z zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func allBytesIn(value string, allowed string) bool {
	for i := 0; i < len(value); i++ {
		if !strings.Contains(allowed, string(value[i])) {
			return false
		}
	}
	return true
}
