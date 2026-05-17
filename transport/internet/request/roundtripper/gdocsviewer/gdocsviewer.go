package gdocsviewer

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	gonet "net"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
	"github.com/v2fly/v2ray-core/v5/transport/internet/transportcommon"
)

const (
	defaultViewerURL            = "https://docs.google.com/viewer"
	defaultTextURL              = "https://drive.google.com/viewerng/text"
	defaultPathPrefix           = "/gdocsviewer"
	defaultMaxViewerBodyBytes   = 32 * 1024 * 1024
	defaultMinRequestIntervalMs = 100
	defaultMaxRequestBytes      = 1100
	defaultMaxResponseBytes     = 64 * 1024
	encryptedRequestVersion     = byte(1)
	responseFrameSuccess        = byte(0)
	responseFrameError          = byte(1)
)

func newClient(ctx context.Context, config *ClientConfig) request.RoundTripperClient {
	return &client{ctx: ctx, config: config}
}

type client struct {
	ctx      context.Context
	httpRTT  http.RoundTripper
	config   *ClientConfig
	assembly request.TransportClientAssembly

	requestLock sync.Mutex
	lastRequest time.Time
}

type unavailableRoundTripper struct{}

func (u unavailableRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, newError("unavailable HTTP transport")
}

func (c *client) OnTransportClientAssemblyReady(assembly request.TransportClientAssembly) {
	c.assembly = assembly
}

func (c *client) RoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	var streamingWriter io.Writer
	for _, v := range opts {
		if streamingResp, ok := v.(request.OptionSupportsStreamingResponse); ok {
			streamingWriter = streamingResp.GetResponseWriter()
		}
	}

	if c.httpRTT == nil {
		c.initHTTPRoundTripper(ctx)
	}

	originRequestURL, err := c.buildOriginRequestURL(req)
	if err != nil {
		return request.Response{}, err
	}

	viewerURL, err := buildViewerURL(clientViewerURL(c.config), originRequestURL)
	if err != nil {
		return request.Response{}, err
	}
	viewerReq, err := http.NewRequestWithContext(ctx, http.MethodGet, viewerURL, nil)
	if err != nil {
		return request.Response{}, err
	}
	c.applyViewerHeaders(viewerReq)

	viewerResp, err := c.doHTTPRoundTrip(ctx, viewerReq)
	if err != nil {
		return request.Response{}, err
	}
	viewerBody, err := readLimitedHTTPBody(viewerResp, int64(clientMaxViewerBodyBytes(c.config)), "viewer")
	if err != nil {
		return request.Response{}, err
	}

	docID, err := extractDocumentID(viewerBody)
	if err != nil {
		return request.Response{}, err
	}

	textURL, err := buildTextURL(clientTextURL(c.config), docID)
	if err != nil {
		return request.Response{}, err
	}
	textReq, err := http.NewRequestWithContext(ctx, http.MethodGet, textURL, nil)
	if err != nil {
		return request.Response{}, err
	}
	c.applyViewerHeaders(textReq)

	textResp, err := c.doHTTPRoundTrip(ctx, textReq)
	if err != nil {
		return request.Response{}, err
	}
	textBody, err := readLimitedHTTPBody(textResp, int64(clientMaxViewerBodyBytes(c.config)), "viewer text")
	if err != nil {
		return request.Response{}, err
	}

	originBase64, err := decodeViewerTextData(textBody)
	if err != nil {
		return request.Response{}, err
	}
	serverBody, err := decodeBase64Text(originBase64)
	if err != nil {
		return request.Response{}, newError("unable to decode origin response body").Base(err)
	}
	result, err := c.decodeServerBody(serverBody)
	if err != nil {
		return request.Response{}, err
	}

	if streamingWriter != nil {
		if _, err := streamingWriter.Write(result); err != nil {
			return request.Response{}, err
		}
		return request.Response{}, nil
	}
	return request.Response{Data: result}, nil
}

func (c *client) initHTTPRoundTripper(ctx context.Context) {
	var backdrop http.RoundTripper = unavailableRoundTripper{}
	if c.config != nil && c.config.AllowHttp {
		backdrop = &http.Transport{
			DialContext: func(_ context.Context, network, addr string) (gonet.Conn, error) {
				return c.assembly.AutoImplDialer().Dial(ctx)
			},
			DialTLSContext: func(_ context.Context, network, addr string) (gonet.Conn, error) {
				return nil, newError("unexpected TLS dial for HTTP request")
			},
		}
	}
	c.httpRTT = transportcommon.NewALPNAwareHTTPRoundTripperWithH2Pool(c.ctx, func(ctx context.Context, addr string) (gonet.Conn, error) {
		return c.assembly.AutoImplDialer().Dial(ctx)
	}, backdrop, int(c.config.GetH2PoolSize()))
}

func (c *client) applyViewerHeaders(req *http.Request) {
	if c.config == nil {
		return
	}
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}
	if c.config.ViewerHostHeader != "" {
		req.Host = c.config.ViewerHostHeader
	}
}

func (c *client) doHTTPRoundTrip(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.waitForRequestSlot(ctx); err != nil {
		return nil, err
	}
	return c.httpRTT.RoundTrip(req)
}

func (c *client) waitForRequestSlot(ctx context.Context) error {
	interval := time.Duration(clientMinRequestIntervalMs(c.config)) * time.Millisecond
	if interval <= 0 {
		return nil
	}

	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	if !c.lastRequest.IsZero() {
		wait := interval - time.Since(c.lastRequest)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	c.lastRequest = time.Now()
	return nil
}

func (c *client) buildOriginRequestURL(req request.Request) (string, error) {
	if c.config == nil || c.config.OriginUrl == "" {
		return "", newError("origin_url is required")
	}
	originURL, err := renderOriginURL(c.config)
	if err != nil {
		return "", err
	}
	if len(c.config.SharedKey) == 0 {
		nonce, err := randomURLToken(12)
		if err != nil {
			return "", err
		}
		return originURL + "/r/" +
			base64.RawURLEncoding.EncodeToString(req.ConnectionTag) + "/" +
			base64.RawURLEncoding.EncodeToString(req.Data) + "/" +
			nonce + ".txt", nil
	}

	combined, err := encryptClientRequest(c.config.SharedKey, req.ConnectionTag, req.Data)
	if err != nil {
		return "", err
	}
	return originURL + "/t/" + base64.RawURLEncoding.EncodeToString(combined) + ".log", nil
}

func (c *client) decodeServerBody(body []byte) ([]byte, error) {
	if c.config == nil || len(c.config.SharedKey) == 0 {
		return body, nil
	}
	frame, err := decryptResponseFrame(c.config.SharedKey, body)
	if err != nil {
		return nil, err
	}
	if len(frame) == 0 {
		return nil, newError("empty encrypted response frame")
	}
	switch frame[0] {
	case responseFrameSuccess:
		return frame[1:], nil
	case responseFrameError:
		return nil, newError("server returned error: ", string(frame[1:]))
	default:
		return nil, newError("unknown encrypted response frame type: ", frame[0])
	}
}

func clientViewerURL(config *ClientConfig) string {
	if config != nil && config.ViewerUrl != "" {
		return config.ViewerUrl
	}
	return defaultViewerURL
}

func clientTextURL(config *ClientConfig) string {
	if config != nil && config.TextUrl != "" {
		return config.TextUrl
	}
	return defaultTextURL
}

func clientMaxViewerBodyBytes(config *ClientConfig) int {
	if config != nil && config.MaxViewerBodyBytes > 0 {
		return int(config.MaxViewerBodyBytes)
	}
	return defaultMaxViewerBodyBytes
}

func clientMinRequestIntervalMs(config *ClientConfig) int {
	if config != nil && config.MinRequestIntervalMs > 0 {
		return int(config.MinRequestIntervalMs)
	}
	return defaultMinRequestIntervalMs
}

func buildViewerURL(viewerURL string, originURL string) (string, error) {
	parsed, err := neturl.Parse(viewerURL)
	if err != nil {
		return "", newError("invalid viewer_url").Base(err)
	}
	query := parsed.Query()
	query.Set("url", originURL)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func buildTextURL(textURL string, docID string) (string, error) {
	parsed, err := neturl.Parse(textURL)
	if err != nil {
		return "", newError("invalid text_url").Base(err)
	}
	query := parsed.Query()
	query.Set("id", docID)
	if query.Get("page") == "" {
		query.Set("page", "0")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func readLimitedHTTPBody(resp *http.Response, maxBytes int64, label string) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, newError(label, " returned non-200 response: ", resp.Status)
	}
	body, err := readLimited(resp.Body, maxBytes)
	if err != nil {
		return nil, newError("unable to read ", label, " body").Base(err)
	}
	return body, nil
}

func readLimited(reader io.Reader, maxBytes int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, newError("body exceeds limit: ", maxBytes)
	}
	return body, nil
}

func extractDocumentID(body []byte) (string, error) {
	original := string(body)
	candidates := []string{
		original,
		decodeKnownURLEscapes(original),
		decodeKnownJSEscapes(original),
		decodeKnownURLEscapes(decodeKnownJSEscapes(original)),
		decodeKnownJSEscapes(decodeKnownURLEscapes(original)),
	}
	for _, candidate := range candidates {
		if id, ok := findDocumentID(candidate); ok {
			return id, nil
		}
	}
	return "", newError("unable to find Google Docs Viewer text document id")
}

func decodeKnownURLEscapes(s string) string {
	replacer := strings.NewReplacer(
		"%2F", "/", "%2f", "/",
		"%3F", "?", "%3f", "?",
		"%3D", "=", "%3d", "=",
		"%26", "&",
	)
	return replacer.Replace(s)
}

func decodeKnownJSEscapes(s string) string {
	replacer := strings.NewReplacer(
		`\/`, `/`,
		`\x2F`, `/`, `\x2f`, `/`, `\u002F`, `/`, `\u002f`, `/`,
		`\x3F`, `?`, `\x3f`, `?`, `\u003F`, `?`, `\u003f`, `?`,
		`\x3D`, `=`, `\x3d`, `=`, `\u003D`, `=`, `\u003d`, `=`,
		`\x26`, `&`, `\u0026`, `&`,
	)
	return replacer.Replace(s)
}

func findDocumentID(s string) (string, bool) {
	for _, marker := range []string{"viewerng/text?id=", "text?id="} {
		if id, ok := findDocumentIDAfterMarker(s, marker); ok {
			return id, true
		}
	}
	return "", false
}

func findDocumentIDAfterMarker(s string, marker string) (string, bool) {
	pos := strings.Index(s, marker)
	if pos < 0 {
		return "", false
	}
	start := pos + len(marker)
	end := start
	for end < len(s) {
		switch s[end] {
		case '&', '"', '\'', '\\', '<', '>', ' ', '\t', '\r', '\n':
			goto done
		default:
			end++
		}
	}
done:
	if end == start {
		return "", false
	}
	id := s[start:end]
	if decoded, err := neturl.QueryUnescape(id); err == nil {
		id = decoded
	}
	return id, true
}

type viewerTextResponse struct {
	Mimetype string `json:"mimetype"`
	Data     string `json:"data"`
}

func decodeViewerTextData(body []byte) ([]byte, error) {
	trimmed := strings.TrimSpace(string(body))
	trimmed = stripAntiXSSIPrefix(trimmed)
	if strings.HasPrefix(trimmed, "{") {
		var response viewerTextResponse
		if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
			return nil, newError("unable to parse viewer text JSON").Base(err)
		}
		if response.Data == "" {
			return nil, newError("viewer text JSON is missing data")
		}
		return []byte(response.Data), nil
	}
	return []byte(trimmed), nil
}

func stripAntiXSSIPrefix(s string) string {
	const prefix = ")]}'"
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, prefix) {
		s = strings.TrimSpace(strings.TrimPrefix(s, prefix))
		if strings.HasPrefix(s, ",") {
			s = strings.TrimSpace(strings.TrimPrefix(s, ","))
		}
	}
	return s
}

func decodeBase64Text(data []byte) ([]byte, error) {
	text := strings.TrimSpace(string(data))
	if decoded, err := base64.StdEncoding.DecodeString(text); err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(text)
}

func encryptClientRequest(key []byte, session []byte, payload []byte) ([]byte, error) {
	if len(session) > 0xffff {
		return nil, newError("session tag is too long")
	}
	frame := make([]byte, 1+2+len(session)+len(payload))
	frame[0] = encryptedRequestVersion
	binary.BigEndian.PutUint16(frame[1:3], uint16(len(session)))
	copy(frame[3:], session)
	copy(frame[3+len(session):], payload)
	return encryptAEAD(key, frame)
}

func decryptClientRequest(key []byte, combined []byte) (session []byte, payload []byte, err error) {
	frame, err := decryptAEAD(key, combined)
	if err != nil {
		return nil, nil, err
	}
	if len(frame) < 3 {
		return nil, nil, newError("encrypted request frame is too short")
	}
	if frame[0] != encryptedRequestVersion {
		return nil, nil, newError("unknown encrypted request frame version: ", frame[0])
	}
	sessionLength := int(binary.BigEndian.Uint16(frame[1:3]))
	if len(frame) < 3+sessionLength {
		return nil, nil, newError("encrypted request frame has invalid session length")
	}
	session = make([]byte, sessionLength)
	copy(session, frame[3:3+sessionLength])
	payload = make([]byte, len(frame)-3-sessionLength)
	copy(payload, frame[3+sessionLength:])
	return session, payload, nil
}

func encryptResponseFrame(key []byte, frame []byte) ([]byte, error) {
	return encryptAEAD(key, frame)
}

func decryptResponseFrame(key []byte, combined []byte) ([]byte, error) {
	return decryptAEAD(key, combined)
}

func encryptAEAD(key []byte, plaintext []byte) ([]byte, error) {
	aead, err := newAES256GCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, newError("unable to generate nonce").Base(err)
	}
	sealed := aead.Seal(nil, nonce, plaintext, nil)
	combined := make([]byte, 0, len(nonce)+len(sealed))
	combined = append(combined, nonce...)
	combined = append(combined, sealed...)
	return combined, nil
}

func decryptAEAD(key []byte, combined []byte) ([]byte, error) {
	aead, err := newAES256GCM(key)
	if err != nil {
		return nil, err
	}
	if len(combined) < aead.NonceSize()+aead.Overhead() {
		return nil, newError("encrypted blob is too short")
	}
	nonce := combined[:aead.NonceSize()]
	ciphertext := combined[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, newError("unable to decrypt encrypted blob").Base(err)
	}
	return plaintext, nil
}

func newAES256GCM(key []byte) (cipher.AEAD, error) {
	if len(key) != 32 {
		return nil, newError("shared_key must be 32 bytes for AES-256-GCM")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, newError("unable to create AES cipher").Base(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, newError("unable to create AES-GCM AEAD").Base(err)
	}
	return aead, nil
}

func randomURLToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", newError("unable to generate URL token").Base(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func newServer(ctx context.Context, config *ServerConfig) request.RoundTripperServer {
	return &server{ctx: ctx, config: config}
}

type server struct {
	ctx      context.Context
	listener net.Listener
	assembly request.TransportServerAssembly
	config   *ServerConfig
}

func (s *server) OnTransportServerAssemblyReady(assembly request.TransportServerAssembly) {
	s.assembly = assembly
}

func (s *server) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	s.onRequest(writer, r)
}

func (s *server) onRequest(writer http.ResponseWriter, httpReq *http.Request) {
	if httpReq.Method != http.MethodGet {
		http.NotFound(writer, httpReq)
		return
	}
	tail, ok := pathTail(httpReq.URL.Path, serverPathPrefix(s.config))
	if !ok {
		http.NotFound(writer, httpReq)
		return
	}
	if len(s.config.GetSharedKey()) == 0 {
		s.handlePlainRequest(writer, httpReq, tail)
		return
	}
	s.handleEncryptedRequest(writer, httpReq, tail)
}

func (s *server) handlePlainRequest(writer http.ResponseWriter, httpReq *http.Request, tail string) {
	if !strings.HasPrefix(tail, "/r/") {
		http.NotFound(writer, httpReq)
		return
	}
	parts := strings.Split(strings.TrimPrefix(tail, "/r/"), "/")
	if len(parts) != 3 || !strings.HasSuffix(parts[2], ".txt") || strings.TrimSuffix(parts[2], ".txt") == "" {
		http.NotFound(writer, httpReq)
		return
	}
	session, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		http.NotFound(writer, httpReq)
		return
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		http.NotFound(writer, httpReq)
		return
	}
	if len(payload) > serverMaxRequestBytes(s.config) {
		newError("plaintext request exceeds max_request_bytes").AtInfo().WriteToLog()
		http.NotFound(writer, httpReq)
		return
	}
	response, err := s.roundTrip(s.roundTripContext(), request.Request{Data: payload, ConnectionTag: session})
	if err != nil {
		newError("unable to process plaintext request").Base(err).AtInfo().WriteToLog()
		http.NotFound(writer, httpReq)
		return
	}
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = io.WriteString(writer, base64.StdEncoding.EncodeToString(response))
	if err != nil {
		newError("unable to write plaintext response").Base(err).AtInfo().WriteToLog()
	}
}

func (s *server) handleEncryptedRequest(writer http.ResponseWriter, httpReq *http.Request, tail string) {
	if !strings.HasPrefix(tail, "/t/") {
		http.NotFound(writer, httpReq)
		return
	}
	combinedText := strings.TrimPrefix(tail, "/t/")
	if !strings.HasSuffix(combinedText, ".log") {
		http.NotFound(writer, httpReq)
		return
	}
	combinedText = strings.TrimSuffix(combinedText, ".log")
	combined, err := base64.RawURLEncoding.DecodeString(combinedText)
	if err != nil {
		s.writeEncryptedError(writer, "invalid encrypted request encoding")
		return
	}
	session, payload, err := decryptClientRequest(s.config.SharedKey, combined)
	if err != nil {
		s.writeEncryptedError(writer, err.Error())
		return
	}
	if len(payload) > serverMaxRequestBytes(s.config) {
		s.writeEncryptedError(writer, "request exceeds max_request_bytes")
		return
	}
	response, err := s.roundTrip(s.roundTripContext(), request.Request{Data: payload, ConnectionTag: session})
	if err != nil {
		s.writeEncryptedError(writer, err.Error())
		return
	}
	frame := make([]byte, 1+len(response))
	frame[0] = responseFrameSuccess
	copy(frame[1:], response)
	encrypted, err := encryptResponseFrame(s.config.SharedKey, frame)
	if err != nil {
		s.writeEncryptedError(writer, err.Error())
		return
	}
	writeBase64Text(writer, encrypted)
}

func (s *server) writeEncryptedError(writer http.ResponseWriter, message string) {
	frame := append([]byte{responseFrameError}, []byte(message)...)
	encrypted, err := encryptResponseFrame(s.config.SharedKey, frame)
	if err != nil {
		newError("unable to encrypt error response").Base(err).AtInfo().WriteToLog()
		http.Error(writer, "encrypted error unavailable", http.StatusInternalServerError)
		return
	}
	writeBase64Text(writer, encrypted)
}

func (s *server) roundTrip(ctx context.Context, req request.Request) ([]byte, error) {
	resp, err := s.assembly.TripperReceiver().OnRoundTrip(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) > serverMaxResponseBytes(s.config) {
		return nil, newError("response exceeds max_response_bytes")
	}
	return resp.Data, nil
}

func (s *server) roundTripContext() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func writeBase64Text(writer http.ResponseWriter, data []byte) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := io.WriteString(writer, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		newError("unable to write response").Base(err).AtInfo().WriteToLog()
	}
}

func pathTail(path string, prefix string) (string, bool) {
	if prefix == "/" {
		return path, strings.HasPrefix(path, "/")
	}
	if path == prefix {
		return "", true
	}
	if strings.HasPrefix(path, prefix+"/") {
		return strings.TrimPrefix(path, prefix), true
	}
	return "", false
}

func serverPathPrefix(config *ServerConfig) string {
	if config != nil && config.PathPrefix != "" {
		prefix := config.PathPrefix
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		prefix = strings.TrimRight(prefix, "/")
		if prefix == "" {
			return "/"
		}
		return prefix
	}
	return defaultPathPrefix
}

func serverMaxRequestBytes(config *ServerConfig) int {
	if config != nil && config.MaxRequestBytes > 0 {
		return int(config.MaxRequestBytes)
	}
	return defaultMaxRequestBytes
}

func serverMaxResponseBytes(config *ServerConfig) int {
	if config != nil && config.MaxResponseBytes > 0 {
		return int(config.MaxResponseBytes)
	}
	return defaultMaxResponseBytes
}

func (s *server) Start() error {
	listener, err := s.assembly.AutoImplListener().Listen(s.ctx)
	if err != nil {
		return newError("unable to create a listener for gdocsviewer server").Base(err)
	}
	s.listener = listener
	go func() {
		httpServer := http.Server{
			ReadHeaderTimeout: 240 * time.Second,
			ReadTimeout:       240 * time.Second,
			WriteTimeout:      240 * time.Second,
			IdleTimeout:       240 * time.Second,
			Handler:           s,
		}
		err := httpServer.Serve(s.listener)
		if err != nil {
			newError("unable to serve listener for gdocsviewer server").Base(err).WriteToLog()
		}
	}()
	return nil
}

func (s *server) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		clientConfig, ok := config.(*ClientConfig)
		if !ok {
			return nil, newError("not a ClientConfig")
		}
		return newClient(ctx, clientConfig), nil
	}))
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		serverConfig, ok := config.(*ServerConfig)
		if !ok {
			return nil, newError("not a ServerConfig")
		}
		return newServer(ctx, serverConfig), nil
	}))
}
