package encoding

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"hash/fnv"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/bitmask"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/crypto"
	"github.com/ghxhy/v2ray-core/v5/common/drain"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/task"
	"github.com/ghxhy/v2ray-core/v5/proxy/vmess"
	vmessaead "github.com/ghxhy/v2ray-core/v5/proxy/vmess/aead"
)

type sessionID struct {
	user  [16]byte
	key   [16]byte
	nonce [16]byte
}

// SessionHistory keeps track of historical session ids, to prevent replay attacks.
type SessionHistory struct {
	sync.RWMutex
	cache map[sessionID]time.Time
	task  *task.Periodic
}

// NewSessionHistory creates a new SessionHistory object.
func NewSessionHistory() *SessionHistory {
	h := &SessionHistory{
		cache: make(map[sessionID]time.Time, 128),
	}
	h.task = &task.Periodic{
		Interval: time.Second * 30,
		Execute:  h.removeExpiredEntries,
	}
	return h
}

// Close implements common.Closable.
func (h *SessionHistory) Close() error {
	return h.task.Close()
}

func (h *SessionHistory) addIfNotExits(session sessionID) bool {
	h.Lock()

	if expire, found := h.cache[session]; found && expire.After(time.Now()) {
		h.Unlock()
		return false
	}

	h.cache[session] = time.Now().Add(time.Minute * 3)
	h.Unlock()
	common.Must(h.task.Start())
	return true
}

func (h *SessionHistory) removeExpiredEntries() error {
	now := time.Now()

	h.Lock()
	defer h.Unlock()

	if len(h.cache) == 0 {
		return newError("nothing to do")
	}

	for session, expire := range h.cache {
		if expire.Before(now) {
			delete(h.cache, session)
		}
	}

	if len(h.cache) == 0 {
		h.cache = make(map[sessionID]time.Time, 128)
	}

	return nil
}

// ServerSession keeps information for a session in VMess server.
type ServerSession struct {
	userValidator   *vmess.TimedUserValidator
	sessionHistory  *SessionHistory
	requestBodyKey  [16]byte
	requestBodyIV   [16]byte
	responseBodyKey [16]byte
	responseBodyIV  [16]byte
	responseWriter  io.Writer
	responseHeader  byte

	isAEADRequest bool

	isAEADForced bool
}

// NewServerSession creates a new ServerSession, using the given UserValidator.
// The ServerSession instance doesn't take ownership of the validator.
func NewServerSession(validator *vmess.TimedUserValidator, sessionHistory *SessionHistory) *ServerSession {
	return &ServerSession{
		userValidator:  validator,
		sessionHistory: sessionHistory,
	}
}

// SetAEADForced sets isAEADForced for a ServerSession.
func (s *ServerSession) SetAEADForced(isAEADForced bool) {
	s.isAEADForced = isAEADForced
}

func parseSecurityType(b byte) protocol.SecurityType {
	if _, f := protocol.SecurityType_name[int32(b)]; f {
		st := protocol.SecurityType(b)
		// For backward compatibility.
		if st == protocol.SecurityType_UNKNOWN {
			st = protocol.SecurityType_LEGACY
		}
		return st
	}
	return protocol.SecurityType_UNKNOWN
}

// DecodeRequestHeader decodes and returns (if successful) a RequestHeader from an input stream.
func (s *ServerSession) DecodeRequestHeader(reader io.Reader) (*protocol.RequestHeader, error) {
	buffer := buf.New()

	drainer, err := drain.NewBehaviorSeedLimitedDrainer(int64(s.userValidator.GetBehaviorSeed()), 16+38, 3266, 64)
	if err != nil {
		return nil, newError("failed to initialize drainer").Base(err)
	}

	drainConnection := func(e error) error {
		// We read a deterministic generated length of data before closing the connection to offset padding read pattern
		drainer.AcknowledgeReceive(int(buffer.Len()))
		return drain.WithError(drainer, reader, e)
	}

	defer func() {
		buffer.Release()
	}()

	if _, err := buffer.ReadFullFrom(reader, protocol.IDBytesLen); err != nil {
		return nil, newError("failed to read request header").Base(err)
	}

	var decryptor io.Reader
	var vmessAccount *vmess.MemoryAccount

	user, foundAEAD, errorAEAD := s.userValidator.GetAEAD(buffer.Bytes())

	var fixedSizeAuthID [16]byte
	copy(fixedSizeAuthID[:], buffer.Bytes())

	switch {
	case foundAEAD:
		vmessAccount = user.Account.(*vmess.MemoryAccount)
		var fixedSizeCmdKey [16]byte
		copy(fixedSizeCmdKey[:], vmessAccount.ID.CmdKey())
		aeadData, shouldDrain, bytesRead, errorReason := vmessaead.OpenVMessAEADHeader(fixedSizeCmdKey, fixedSizeAuthID, reader)
		if errorReason != nil {
			if shouldDrain {
				drainer.AcknowledgeReceive(bytesRead)
				return nil, drainConnection(newError("AEAD read failed").Base(errorReason))
			}
			return nil, drainConnection(newError("AEAD read failed, drain skipped").Base(errorReason))
		}
		decryptor = bytes.NewReader(aeadData)
		s.isAEADRequest = true

	case errorAEAD == vmessaead.ErrNotFound:
		userLegacy, timestamp, valid, userValidationError := s.userValidator.Get(buffer.Bytes())
		if !valid || userValidationError != nil {
			return nil, drainConnection(newError("invalid user").Base(userValidationError))
		}
		if s.isAEADForced {
			return nil, drainConnection(newError("invalid user: VMessAEAD is enforced and a non VMessAEAD connection is received. You can still disable this security feature with environment variable v2ray.vmess.aead.forced = false . You will not be able to enable legacy header workaround in the future."))
		}
		if s.userValidator.ShouldShowLegacyWarn() {
			newError("Critical Warning: potentially invalid user: a non VMessAEAD connection is received. From 2022 Jan 1st, this kind of connection will be rejected by default. You should update or replace your client software now. This message will not be shown for further violation on this inbound.").AtWarning().WriteToLog()
		}
		user = userLegacy
		iv := hashTimestamp(md5.New(), timestamp)
		vmessAccount = userLegacy.Account.(*vmess.MemoryAccount)

		aesStream := crypto.NewAesDecryptionStream(vmessAccount.ID.CmdKey(), iv)
		decryptor = crypto.NewCryptionReader(aesStream, reader)

	default:
		return nil, drainConnection(newError("invalid user").Base(errorAEAD))
	}

	drainer.AcknowledgeReceive(int(buffer.Len()))
	buffer.Clear()
	if _, err := buffer.ReadFullFrom(decryptor, 38); err != nil {
		return nil, newError("failed to read request header").Base(err)
	}

	request := &protocol.RequestHeader{
		User:    user,
		Version: buffer.Byte(0),
	}

	copy(s.requestBodyIV[:], buffer.BytesRange(1, 17))   // 16 bytes
	copy(s.requestBodyKey[:], buffer.BytesRange(17, 33)) // 16 bytes
	var sid sessionID
	copy(sid.user[:], vmessAccount.ID.Bytes())
	sid.key = s.requestBodyKey
	sid.nonce = s.requestBodyIV
	if !s.sessionHistory.addIfNotExits(sid) {
		if !s.isAEADRequest {
			drainErr := s.userValidator.BurnTaintFuse(fixedSizeAuthID[:])
			if drainErr != nil {
				return nil, drainConnection(newError("duplicated session id, possibly under replay attack, and failed to taint userHash").Base(drainErr))
			}
			return nil, drainConnection(newError("duplicated session id, possibly under replay attack, userHash tainted"))
		}
		return nil, newError("duplicated session id, possibly under replay attack, but this is a AEAD request")
	}

	s.responseHeader = buffer.Byte(33)             // 1 byte
	request.Option = bitmask.Byte(buffer.Byte(34)) // 1 byte
	paddingLen := int(buffer.Byte(35) >> 4)
	request.Security = parseSecurityType(buffer.Byte(35) & 0x0F)
	// 1 bytes reserved
	request.Command = protocol.RequestCommand(buffer.Byte(37))

	switch request.Command {
	case protocol.RequestCommandMux:
		request.Address = net.DomainAddress("v1.mux.cool")
		request.Port = 0

	case protocol.RequestCommandTCP, protocol.RequestCommandUDP:
		if addr, port, err := addrParser.ReadAddressPort(buffer, decryptor); err == nil {
			request.Address = addr
			request.Port = port
		}
	}

	if paddingLen > 0 {
		if _, err := buffer.ReadFullFrom(decryptor, int32(paddingLen)); err != nil {
			if !s.isAEADRequest {
				burnErr := s.userValidator.BurnTaintFuse(fixedSizeAuthID[:])
				if burnErr != nil {
					return nil, newError("failed to read padding, failed to taint userHash").Base(burnErr).Base(err)
				}
				return nil, newError("failed to read padding, userHash tainted").Base(err)
			}
			return nil, newError("failed to read padding").Base(err)
		}
	}

	if _, err := buffer.ReadFullFrom(decryptor, 4); err != nil {
		if !s.isAEADRequest {
			burnErr := s.userValidator.BurnTaintFuse(fixedSizeAuthID[:])
			if burnErr != nil {
				return nil, newError("failed to read checksum, failed to taint userHash").Base(burnErr).Base(err)
			}
			return nil, newError("failed to read checksum, userHash tainted").Base(err)
		}
		return nil, newError("failed to read checksum").Base(err)
	}

	fnv1a := fnv.New32a()
	common.Must2(fnv1a.Write(buffer.BytesTo(-4)))
	actualHash := fnv1a.Sum32()
	expectedHash := binary.BigEndian.Uint32(buffer.BytesFrom(-4))

	if actualHash != expectedHash {
		if !s.isAEADRequest {
			Autherr := newError("invalid auth, legacy userHash tainted")
			burnErr := s.userValidator.BurnTaintFuse(fixedSizeAuthID[:])
			if burnErr != nil {
				Autherr = newError("invalid auth, can't taint legacy userHash").Base(burnErr)
			}
			// It is possible that we are under attack described in https://github.com/v2ray/v2ray-core/issues/2523
			return nil, drainConnection(Autherr)
		}
		return nil, newError("invalid auth, but this is a AEAD request")
	}

	if request.Address == nil {
		return nil, newError("invalid remote address")
	}

	if request.Security == protocol.SecurityType_UNKNOWN || request.Security == protocol.SecurityType_AUTO {
		return nil, newError("unknown security type: ", request.Security)
	}

	return request, nil
}

// DecodeRequestBody returns Reader from which caller can fetch decrypted body.
func (s *ServerSession) DecodeRequestBody(request *protocol.RequestHeader, reader io.Reader) (buf.Reader, error) {
	var sizeParser crypto.ChunkSizeDecoder = crypto.PlainChunkSizeParser{}
	if request.Option.Has(protocol.RequestOptionChunkMasking) {
		sizeParser = NewShakeSizeParser(s.requestBodyIV[:])
	}
	var padding crypto.PaddingLengthGenerator
	if request.Option.Has(protocol.RequestOptionGlobalPadding) {
		var ok bool
		padding, ok = sizeParser.(crypto.PaddingLengthGenerator)
		if !ok {
			return nil, newError("invalid option: RequestOptionGlobalPadding")
		}
	}

	switch request.Security {
	case protocol.SecurityType_NONE:
		if request.Option.Has(protocol.RequestOptionChunkStream) {
			if request.Command.TransferType() == protocol.TransferTypeStream {
				return crypto.NewChunkStreamReader(sizeParser, reader), nil
			}

			auth := &crypto.AEADAuthenticator{
				AEAD:                    new(NoOpAuthenticator),
				NonceGenerator:          crypto.GenerateEmptyBytes(),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			return crypto.NewAuthenticationReader(auth, sizeParser, reader, protocol.TransferTypePacket, padding), nil
		}
		return buf.NewReader(reader), nil

	case protocol.SecurityType_LEGACY:
		aesStream := crypto.NewAesDecryptionStream(s.requestBodyKey[:], s.requestBodyIV[:])
		cryptionReader := crypto.NewCryptionReader(aesStream, reader)
		if request.Option.Has(protocol.RequestOptionChunkStream) {
			auth := &crypto.AEADAuthenticator{
				AEAD:                    new(FnvAuthenticator),
				NonceGenerator:          crypto.GenerateEmptyBytes(),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			return crypto.NewAuthenticationReader(auth, sizeParser, cryptionReader, request.Command.TransferType(), padding), nil
		}
		return buf.NewReader(cryptionReader), nil

	case protocol.SecurityType_AES128_GCM:
		aead := crypto.NewAesGcm(s.requestBodyKey[:])
		auth := &crypto.AEADAuthenticator{
			AEAD:                    aead,
			NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
			AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
		}
		if request.Option.Has(protocol.RequestOptionAuthenticatedLength) {
			AuthenticatedLengthKey := vmessaead.KDF16(s.requestBodyKey[:], "auth_len")
			AuthenticatedLengthKeyAEAD := crypto.NewAesGcm(AuthenticatedLengthKey)

			lengthAuth := &crypto.AEADAuthenticator{
				AEAD:                    AuthenticatedLengthKeyAEAD,
				NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			sizeParser = NewAEADSizeParser(lengthAuth)
		}
		return crypto.NewAuthenticationReader(auth, sizeParser, reader, request.Command.TransferType(), padding), nil

	case protocol.SecurityType_CHACHA20_POLY1305:
		aead, _ := chacha20poly1305.New(GenerateChacha20Poly1305Key(s.requestBodyKey[:]))

		auth := &crypto.AEADAuthenticator{
			AEAD:                    aead,
			NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
			AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
		}
		if request.Option.Has(protocol.RequestOptionAuthenticatedLength) {
			AuthenticatedLengthKey := vmessaead.KDF16(s.requestBodyKey[:], "auth_len")
			AuthenticatedLengthKeyAEAD, err := chacha20poly1305.New(GenerateChacha20Poly1305Key(AuthenticatedLengthKey))
			common.Must(err)

			lengthAuth := &crypto.AEADAuthenticator{
				AEAD:                    AuthenticatedLengthKeyAEAD,
				NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			sizeParser = NewAEADSizeParser(lengthAuth)
		}
		return crypto.NewAuthenticationReader(auth, sizeParser, reader, request.Command.TransferType(), padding), nil

	default:
		return nil, newError("invalid option: Security")
	}
}

// EncodeResponseHeader writes encoded response header into the given writer.
func (s *ServerSession) EncodeResponseHeader(header *protocol.ResponseHeader, writer io.Writer) {
	var encryptionWriter io.Writer
	if !s.isAEADRequest {
		s.responseBodyKey = md5.Sum(s.requestBodyKey[:])
		s.responseBodyIV = md5.Sum(s.requestBodyIV[:])
	} else {
		BodyKey := sha256.Sum256(s.requestBodyKey[:])
		copy(s.responseBodyKey[:], BodyKey[:16])
		BodyIV := sha256.Sum256(s.requestBodyIV[:])
		copy(s.responseBodyIV[:], BodyIV[:16])
	}

	aesStream := crypto.NewAesEncryptionStream(s.responseBodyKey[:], s.responseBodyIV[:])
	encryptionWriter = crypto.NewCryptionWriter(aesStream, writer)
	s.responseWriter = encryptionWriter

	aeadEncryptedHeaderBuffer := bytes.NewBuffer(nil)

	if s.isAEADRequest {
		encryptionWriter = aeadEncryptedHeaderBuffer
	}

	common.Must2(encryptionWriter.Write([]byte{s.responseHeader, byte(header.Option)}))
	err := MarshalCommand(header.Command, encryptionWriter)
	if err != nil {
		common.Must2(encryptionWriter.Write([]byte{0x00, 0x00}))
	}

	if s.isAEADRequest {
		aeadResponseHeaderLengthEncryptionKey := vmessaead.KDF16(s.responseBodyKey[:], vmessaead.KDFSaltConstAEADRespHeaderLenKey)
		aeadResponseHeaderLengthEncryptionIV := vmessaead.KDF(s.responseBodyIV[:], vmessaead.KDFSaltConstAEADRespHeaderLenIV)[:12]

		aeadResponseHeaderLengthEncryptionKeyAESBlock := common.Must2(aes.NewCipher(aeadResponseHeaderLengthEncryptionKey)).(cipher.Block)
		aeadResponseHeaderLengthEncryptionAEAD := common.Must2(cipher.NewGCM(aeadResponseHeaderLengthEncryptionKeyAESBlock)).(cipher.AEAD)

		aeadResponseHeaderLengthEncryptionBuffer := bytes.NewBuffer(nil)

		decryptedResponseHeaderLengthBinaryDeserializeBuffer := uint16(aeadEncryptedHeaderBuffer.Len())

		common.Must(binary.Write(aeadResponseHeaderLengthEncryptionBuffer, binary.BigEndian, decryptedResponseHeaderLengthBinaryDeserializeBuffer))

		AEADEncryptedLength := aeadResponseHeaderLengthEncryptionAEAD.Seal(nil, aeadResponseHeaderLengthEncryptionIV, aeadResponseHeaderLengthEncryptionBuffer.Bytes(), nil)
		common.Must2(io.Copy(writer, bytes.NewReader(AEADEncryptedLength)))

		aeadResponseHeaderPayloadEncryptionKey := vmessaead.KDF16(s.responseBodyKey[:], vmessaead.KDFSaltConstAEADRespHeaderPayloadKey)
		aeadResponseHeaderPayloadEncryptionIV := vmessaead.KDF(s.responseBodyIV[:], vmessaead.KDFSaltConstAEADRespHeaderPayloadIV)[:12]

		aeadResponseHeaderPayloadEncryptionKeyAESBlock := common.Must2(aes.NewCipher(aeadResponseHeaderPayloadEncryptionKey)).(cipher.Block)
		aeadResponseHeaderPayloadEncryptionAEAD := common.Must2(cipher.NewGCM(aeadResponseHeaderPayloadEncryptionKeyAESBlock)).(cipher.AEAD)

		aeadEncryptedHeaderPayload := aeadResponseHeaderPayloadEncryptionAEAD.Seal(nil, aeadResponseHeaderPayloadEncryptionIV, aeadEncryptedHeaderBuffer.Bytes(), nil)
		common.Must2(io.Copy(writer, bytes.NewReader(aeadEncryptedHeaderPayload)))
	}
}

// EncodeResponseBody returns a Writer that auto-encrypt content written by caller.
func (s *ServerSession) EncodeResponseBody(request *protocol.RequestHeader, writer io.Writer) (buf.Writer, error) {
	var sizeParser crypto.ChunkSizeEncoder = crypto.PlainChunkSizeParser{}
	if request.Option.Has(protocol.RequestOptionChunkMasking) {
		sizeParser = NewShakeSizeParser(s.responseBodyIV[:])
	}
	var padding crypto.PaddingLengthGenerator
	if request.Option.Has(protocol.RequestOptionGlobalPadding) {
		var ok bool
		padding, ok = sizeParser.(crypto.PaddingLengthGenerator)
		if !ok {
			return nil, newError("invalid option: RequestOptionGlobalPadding")
		}
	}

	switch request.Security {
	case protocol.SecurityType_NONE:
		if request.Option.Has(protocol.RequestOptionChunkStream) {
			if request.Command.TransferType() == protocol.TransferTypeStream {
				return crypto.NewChunkStreamWriter(sizeParser, writer), nil
			}

			auth := &crypto.AEADAuthenticator{
				AEAD:                    new(NoOpAuthenticator),
				NonceGenerator:          crypto.GenerateEmptyBytes(),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			return crypto.NewAuthenticationWriter(auth, sizeParser, writer, protocol.TransferTypePacket, padding), nil
		}
		return buf.NewWriter(writer), nil

	case protocol.SecurityType_LEGACY:
		if request.Option.Has(protocol.RequestOptionChunkStream) {
			auth := &crypto.AEADAuthenticator{
				AEAD:                    new(FnvAuthenticator),
				NonceGenerator:          crypto.GenerateEmptyBytes(),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			return crypto.NewAuthenticationWriter(auth, sizeParser, s.responseWriter, request.Command.TransferType(), padding), nil
		}
		return &buf.SequentialWriter{Writer: s.responseWriter}, nil

	case protocol.SecurityType_AES128_GCM:
		aead := crypto.NewAesGcm(s.responseBodyKey[:])
		auth := &crypto.AEADAuthenticator{
			AEAD:                    aead,
			NonceGenerator:          GenerateChunkNonce(s.responseBodyIV[:], uint32(aead.NonceSize())),
			AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
		}
		if request.Option.Has(protocol.RequestOptionAuthenticatedLength) {
			AuthenticatedLengthKey := vmessaead.KDF16(s.requestBodyKey[:], "auth_len")
			AuthenticatedLengthKeyAEAD := crypto.NewAesGcm(AuthenticatedLengthKey)

			lengthAuth := &crypto.AEADAuthenticator{
				AEAD:                    AuthenticatedLengthKeyAEAD,
				NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			sizeParser = NewAEADSizeParser(lengthAuth)
		}
		return crypto.NewAuthenticationWriter(auth, sizeParser, writer, request.Command.TransferType(), padding), nil

	case protocol.SecurityType_CHACHA20_POLY1305:
		aead, _ := chacha20poly1305.New(GenerateChacha20Poly1305Key(s.responseBodyKey[:]))

		auth := &crypto.AEADAuthenticator{
			AEAD:                    aead,
			NonceGenerator:          GenerateChunkNonce(s.responseBodyIV[:], uint32(aead.NonceSize())),
			AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
		}
		if request.Option.Has(protocol.RequestOptionAuthenticatedLength) {
			AuthenticatedLengthKey := vmessaead.KDF16(s.requestBodyKey[:], "auth_len")
			AuthenticatedLengthKeyAEAD, err := chacha20poly1305.New(GenerateChacha20Poly1305Key(AuthenticatedLengthKey))
			common.Must(err)

			lengthAuth := &crypto.AEADAuthenticator{
				AEAD:                    AuthenticatedLengthKeyAEAD,
				NonceGenerator:          GenerateChunkNonce(s.requestBodyIV[:], uint32(aead.NonceSize())),
				AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
			}
			sizeParser = NewAEADSizeParser(lengthAuth)
		}
		return crypto.NewAuthenticationWriter(auth, sizeParser, writer, request.Command.TransferType(), padding), nil

	default:
		return nil, newError("invalid option: Security")
	}
}
