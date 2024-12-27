package shadowsocks2022

import (
	"bytes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/binary"
	"io"
	"time"

	"github.com/v2fly/struc"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/crypto"
	"github.com/v2fly/v2ray-core/v5/common/dice"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

type TCPRequest struct {
	keyDerivation KeyDerivation
	method        Method

	c2sSalt  RequestSalt
	c2sNonce crypto.BytesGenerator
	c2sAEAD  cipher.AEAD

	s2cSalt  RequestSalt
	s2cNonce crypto.BytesGenerator
	s2cAEAD  cipher.AEAD

	s2cSaltAssert         RequestSalt
	s2cInitialPayloadSize int
}

func (t *TCPRequest) EncodeTCPRequestHeader(effectivePsk []byte,
	eih [][]byte, address DestinationAddress, destPort int, initialPayload []byte, out *buf.Buffer,
) error {
	requestSalt := newRequestSaltWithLength(t.method.GetSessionSubKeyAndSaltLength())
	{
		err := requestSalt.FillAllFrom(cryptoRand.Reader)
		if err != nil {
			return newError("failed to fill salt").Base(err)
		}
	}
	t.c2sSalt = requestSalt
	sessionKey := make([]byte, t.method.GetSessionSubKeyAndSaltLength())
	{
		err := t.keyDerivation.GetSessionSubKey(effectivePsk, requestSalt.Bytes(), sessionKey)
		if err != nil {
			return newError("failed to get session sub key").Base(err)
		}
	}

	aead, err := t.method.GetStreamAEAD(sessionKey)
	if err != nil {
		return newError("failed to get stream AEAD").Base(err)
	}
	t.c2sAEAD = aead
	paddingLength := TCPMinPaddingLength
	if initialPayload == nil {
		initialPayload = []byte{}
		paddingLength += 1 + dice.RollWith(TCPMaxPaddingLength, cryptoRand.Reader)
	}

	variableLengthHeader := &TCPRequestHeader3VariableLength{
		DestinationAddress: address,
		Contents: struct {
			PaddingLength uint16 `struc:"sizeof=Padding"`
			Padding       []byte
		}(struct {
			PaddingLength uint16
			Padding       []byte
		}{
			PaddingLength: uint16(paddingLength),
			Padding:       make([]byte, paddingLength),
		}),
	}
	variableLengthHeaderBuffer := buf.New()
	defer variableLengthHeaderBuffer.Release()
	{
		err := addrParser.WriteAddressPort(variableLengthHeaderBuffer, address, net.Port(destPort))
		if err != nil {
			return newError("failed to write address port").Base(err)
		}
	}
	{
		err := struc.Pack(variableLengthHeaderBuffer, &variableLengthHeader.Contents)
		if err != nil {
			return newError("failed to pack variable length header").Base(err)
		}
	}
	{
		_, err := variableLengthHeaderBuffer.Write(initialPayload)
		if err != nil {
			return newError("failed to write initial payload").Base(err)
		}
	}

	fixedLengthHeader := &TCPRequestHeader2FixedLength{
		Type:         TCPHeaderTypeClientToServerStream,
		Timestamp:    uint64(time.Now().Unix()),
		HeaderLength: uint16(variableLengthHeaderBuffer.Len()),
	}

	fixedLengthHeaderBuffer := buf.New()
	defer fixedLengthHeaderBuffer.Release()
	{
		err := struc.Pack(fixedLengthHeaderBuffer, fixedLengthHeader)
		if err != nil {
			return newError("failed to pack fixed length header").Base(err)
		}
	}
	eihHeader := ExtensibleIdentityHeaders(newAESEIH(0))
	if len(eih) != 0 {
		eihGenerator := newAESEIHGeneratorContainer(len(eih), effectivePsk, eih)
		eihHeaderGenerated, err := eihGenerator.GenerateEIH(t.keyDerivation, t.method, requestSalt.Bytes())
		if err != nil {
			return newError("failed to construct EIH").Base(err)
		}
		eihHeader = eihHeaderGenerated
	}
	preSessionKeyHeader := &TCPRequestHeader1PreSessionKey{
		Salt: requestSalt,
		EIH:  eihHeader,
	}
	preSessionKeyHeaderBuffer := buf.New()
	defer preSessionKeyHeaderBuffer.Release()
	{
		err := struc.Pack(preSessionKeyHeaderBuffer, preSessionKeyHeader)
		if err != nil {
			return newError("failed to pack pre session key header").Base(err)
		}
	}
	requestNonce := crypto.GenerateInitialAEADNonce()
	t.c2sNonce = requestNonce
	{
		n, err := out.Write(preSessionKeyHeaderBuffer.BytesFrom(0))
		if err != nil {
			return newError("failed to write pre session key header").Base(err)
		}
		if int32(n) != preSessionKeyHeaderBuffer.Len() {
			return newError("failed to write pre session key header")
		}
	}
	{
		fixedLengthEncrypted := out.Extend(fixedLengthHeaderBuffer.Len() + int32(aead.Overhead()))
		aead.Seal(fixedLengthEncrypted[:0], requestNonce(), fixedLengthHeaderBuffer.Bytes(), nil)
	}
	{
		variableLengthEncrypted := out.Extend(variableLengthHeaderBuffer.Len() + int32(aead.Overhead()))
		aead.Seal(variableLengthEncrypted[:0], requestNonce(), variableLengthHeaderBuffer.Bytes(), nil)
	}
	return nil
}

func (t *TCPRequest) DecodeTCPResponseHeader(effectivePsk []byte, in io.Reader) error {
	var preSessionKeyHeader TCPResponseHeader1PreSessionKey
	preSessionKeyHeader.Salt = newRequestSaltWithLength(t.method.GetSessionSubKeyAndSaltLength())
	{
		err := struc.Unpack(in, &preSessionKeyHeader)
		if err != nil {
			return newError("failed to unpack pre session key header").Base(err)
		}
	}
	s2cSalt := preSessionKeyHeader.Salt.Bytes()
	t.s2cSalt = preSessionKeyHeader.Salt
	sessionKey := make([]byte, t.method.GetSessionSubKeyAndSaltLength())
	{
		err := t.keyDerivation.GetSessionSubKey(effectivePsk, s2cSalt, sessionKey)
		if err != nil {
			return newError("failed to get session sub key").Base(err)
		}
	}
	aead, err := t.method.GetStreamAEAD(sessionKey)
	if err != nil {
		return newError("failed to get stream AEAD").Base(err)
	}
	t.s2cAEAD = aead

	fixedLengthHeaderEncryptedBuffer := buf.New()
	defer fixedLengthHeaderEncryptedBuffer.Release()
	{
		_, err := fixedLengthHeaderEncryptedBuffer.ReadFullFrom(in, 11+int32(t.method.GetSessionSubKeyAndSaltLength())+int32(aead.Overhead()))
		if err != nil {
			return newError("failed to read fixed length header encrypted").Base(err)
		}
	}
	s2cNonce := crypto.GenerateInitialAEADNonce()
	t.s2cNonce = s2cNonce
	fixedLengthHeaderDecryptedBuffer := buf.New()
	defer fixedLengthHeaderDecryptedBuffer.Release()
	{
		decryptionBuffer := fixedLengthHeaderDecryptedBuffer.Extend(11 + int32(t.method.GetSessionSubKeyAndSaltLength()))
		_, err = aead.Open(decryptionBuffer[:0], s2cNonce(), fixedLengthHeaderEncryptedBuffer.Bytes(), nil)
		if err != nil {
			return newError("failed to decrypt fixed length header").Base(err)
		}
	}
	var fixedLengthHeader TCPResponseHeader2FixedLength
	fixedLengthHeader.RequestSalt = newRequestSaltWithLength(t.method.GetSessionSubKeyAndSaltLength())
	{
		err := struc.Unpack(bytes.NewReader(fixedLengthHeaderDecryptedBuffer.Bytes()), &fixedLengthHeader)
		if err != nil {
			return newError("failed to unpack fixed length header").Base(err)
		}
	}

	if fixedLengthHeader.Type != TCPHeaderTypeServerToClientStream {
		return newError("unexpected TCP header type")
	}
	timeDifference := int64(fixedLengthHeader.Timestamp) - time.Now().Unix()
	if timeDifference < -30 || timeDifference > 30 {
		return newError("timestamp is too far away, timeDifference = ", timeDifference)
	}

	t.s2cSaltAssert = fixedLengthHeader.RequestSalt
	t.s2cInitialPayloadSize = int(fixedLengthHeader.InitialPayloadLength)
	return nil
}

func (t *TCPRequest) CheckC2SConnectionConstraint() error {
	if !bytes.Equal(t.c2sSalt.Bytes(), t.s2cSaltAssert.Bytes()) {
		return newError("c2s salt not equal to s2c salt assert")
	}
	return nil
}

func (t *TCPRequest) CreateClientS2CReader(in io.Reader, initialPayload *buf.Buffer) (buf.Reader, error) {
	AEADAuthenticator := &crypto.AEADAuthenticator{
		AEAD:                    t.s2cAEAD,
		NonceGenerator:          t.s2cNonce,
		AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
	}
	initialPayloadEncrypted := buf.NewWithSize(65535)
	defer initialPayloadEncrypted.Release()
	initialPayloadEncryptedBytes := initialPayloadEncrypted.Extend(int32(t.s2cAEAD.Overhead()) + int32(t.s2cInitialPayloadSize))
	_, err := io.ReadFull(in, initialPayloadEncryptedBytes)
	if err != nil {
		return nil, newError("failed to read initial payload").Base(err)
	}
	initialPayloadBytes := initialPayload.Extend(int32(t.s2cInitialPayloadSize))
	_, err = t.s2cAEAD.Open(initialPayloadBytes[:0], t.s2cNonce(), initialPayloadEncryptedBytes, nil)
	if err != nil {
		return nil, newError("failed to decrypt initial payload").Base(err)
	}
	return crypto.NewAuthenticationReader(AEADAuthenticator, &AEADChunkSizeParser{
		Auth: AEADAuthenticator,
	}, in, protocol.TransferTypeStream, nil), nil
}

func (t *TCPRequest) CreateClientC2SWriter(writer io.Writer) buf.Writer {
	AEADAuthenticator := &crypto.AEADAuthenticator{
		AEAD:                    t.c2sAEAD,
		NonceGenerator:          t.c2sNonce,
		AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
	}
	sizeParser := &crypto.AEADChunkSizeParser{
		Auth: AEADAuthenticator,
	}
	return crypto.NewAuthenticationWriter(AEADAuthenticator, sizeParser, writer, protocol.TransferTypeStream, nil)
}

type AEADChunkSizeParser struct {
	Auth *crypto.AEADAuthenticator
}

func (p *AEADChunkSizeParser) HasConstantOffset() uint16 {
	return uint16(p.Auth.Overhead())
}

func (p *AEADChunkSizeParser) SizeBytes() int32 {
	return 2 + int32(p.Auth.Overhead())
}

func (p *AEADChunkSizeParser) Encode(size uint16, b []byte) []byte {
	binary.BigEndian.PutUint16(b, size-uint16(p.Auth.Overhead()))
	b, err := p.Auth.Seal(b[:0], b[:2])
	common.Must(err)
	return b
}

func (p *AEADChunkSizeParser) Decode(b []byte) (uint16, error) {
	b, err := p.Auth.Open(b[:0], b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}
