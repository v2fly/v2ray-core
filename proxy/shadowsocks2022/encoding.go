package shadowsocks2022

import (
	"bytes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"io"
	"math/rand"
	"time"

	"github.com/lunixbochs/struc"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/crypto"
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
	eih [][]byte, address DestinationAddress, destPort int, initialPayload []byte, Out *buf.Buffer) error {
	requestSalt := newRequestSaltWithLength(t.method.GetSessionSubKeyAndSaltLength())
	{
		err := requestSalt.FillAllFrom(cryptoRand.Reader)
		if err != nil {
			return newError("failed to fill salt").Base(err)
		}
	}
	t.c2sSalt = requestSalt
	var sessionKey = make([]byte, t.method.GetSessionSubKeyAndSaltLength())
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
	var paddingLength = TCPMinPaddingLength
	if initialPayload == nil {
		initialPayload = []byte{}
		paddingLength += rand.Intn(TCPMaxPaddingLength) // TODO INSECURE RANDOM USED
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
	eihGenerator := newAESEIHGeneratorContainer(len(eih), effectivePsk, eih)
	eihHeader, err := eihGenerator.GenerateEIH(t.keyDerivation, t.method, requestSalt.Bytes())
	if err != nil {
		return newError("failed to construct EIH").Base(err)
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
		n, err := Out.Write(preSessionKeyHeaderBuffer.BytesFrom(0))
		if err != nil {
			return newError("failed to write pre session key header").Base(err)
		}
		if int32(n) != preSessionKeyHeaderBuffer.Len() {
			return newError("failed to write pre session key header")
		}
	}
	{
		fixedLengthEncrypted := Out.Extend(fixedLengthHeaderBuffer.Len() + int32(aead.Overhead()))
		aead.Seal(fixedLengthEncrypted[:0], requestNonce(), fixedLengthHeaderBuffer.Bytes(), nil)
	}
	{
		variableLengthEncrypted := Out.Extend(variableLengthHeaderBuffer.Len() + int32(aead.Overhead()))
		aead.Seal(variableLengthEncrypted[:0], requestNonce(), variableLengthHeaderBuffer.Bytes(), nil)
	}
	return nil
}

func (t *TCPRequest) DecodeTCPResponseHeader(effectivePsk []byte, In io.Reader) error {
	var preSessionKeyHeader TCPResponseHeader1PreSessionKey
	preSessionKeyHeader.Salt = newRequestSaltWithLength(t.method.GetSessionSubKeyAndSaltLength())
	{
		err := struc.Unpack(In, &preSessionKeyHeader)
		if err != nil {
			return newError("failed to unpack pre session key header").Base(err)
		}
	}
	var s2cSalt = preSessionKeyHeader.Salt.Bytes()
	t.s2cSalt = preSessionKeyHeader.Salt
	var sessionKey = make([]byte, t.method.GetSessionSubKeyAndSaltLength())
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

	var fixedLengthHeaderEncryptedBuffer = buf.New()
	defer fixedLengthHeaderEncryptedBuffer.Release()
	{
		_, err := fixedLengthHeaderEncryptedBuffer.ReadFullFrom(In, 11+int32(t.method.GetSessionSubKeyAndSaltLength())+int32(aead.Overhead()))
		if err != nil {
			return newError("failed to read fixed length header encrypted").Base(err)
		}
	}
	s2cNonce := crypto.GenerateInitialAEADNonce()
	t.s2cNonce = s2cNonce
	var fixedLengthHeaderDecryptedBuffer = buf.New()
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
	if timeDifference < -60 || timeDifference > 60 {
		return newError("timestamp is too far away")
	}

	t.s2cSaltAssert = fixedLengthHeader.RequestSalt
	t.s2cInitialPayloadSize = int(fixedLengthHeader.InitialPayloadLength)
	return nil
}

func (t *TCPRequest) CheckC2SConnectionConstraint() error {
	if bytes.Compare(t.c2sSalt.Bytes(), t.s2cSaltAssert.Bytes()) != 0 {
		return newError("c2s salt not equal to s2c salt assert")
	}
	return nil
}

func (t *TCPRequest) CreateClientS2CReader(In io.Reader, initialPayload *buf.Buffer) (buf.Reader, error) {
	AEADAuthenticator := &crypto.AEADAuthenticator{
		AEAD:                    t.s2cAEAD,
		NonceGenerator:          t.s2cNonce,
		AdditionalDataGenerator: crypto.GenerateEmptyBytes(),
	}
	initialPayloadEncrypted := buf.NewWithSize(65535)
	defer initialPayloadEncrypted.Release()
	initialPayloadEncryptedBytes := initialPayloadEncrypted.Extend(int32(t.s2cAEAD.Overhead()) + int32(t.s2cInitialPayloadSize))
	_, err := io.ReadFull(In, initialPayloadEncryptedBytes)
	if err != nil {
		return nil, newError("failed to read initial payload").Base(err)
	}
	initialPayloadBytes := initialPayload.Extend(int32(t.s2cInitialPayloadSize))
	_, err = t.s2cAEAD.Open(initialPayloadBytes[:0], t.s2cNonce(), initialPayloadEncryptedBytes, nil)
	if err != nil {
		return nil, newError("failed to decrypt initial payload").Base(err)
	}
	return crypto.NewAuthenticationReader(AEADAuthenticator, &crypto.AEADChunkSizeParser{
		Auth: AEADAuthenticator,
	}, In, protocol.TransferTypeStream, nil), nil
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
