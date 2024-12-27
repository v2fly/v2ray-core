package shadowsocks2022

import (
	"bytes"
	"crypto/cipher"
	"io"

	"github.com/v2fly/struc"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

type AESUDPClientPacketProcessor struct {
	requestSeparateHeaderBlockCipher  cipher.Block
	responseSeparateHeaderBlockCipher cipher.Block
	mainPacketAEAD                    func([]byte) cipher.AEAD
	EIHGenerator                      func([]byte) ExtensibleIdentityHeaders
}

func NewAESUDPClientPacketProcessor(requestSeparateHeaderBlockCipher, responseSeparateHeaderBlockCipher cipher.Block, mainPacketAEAD func([]byte) cipher.AEAD, eih func([]byte) ExtensibleIdentityHeaders) *AESUDPClientPacketProcessor {
	return &AESUDPClientPacketProcessor{
		requestSeparateHeaderBlockCipher:  requestSeparateHeaderBlockCipher,
		responseSeparateHeaderBlockCipher: responseSeparateHeaderBlockCipher,
		mainPacketAEAD:                    mainPacketAEAD,
		EIHGenerator:                      eih,
	}
}

type separateHeader struct {
	SessionID [8]byte
	PacketID  uint64
}

type header struct {
	Type          byte
	TimeStamp     uint64
	PaddingLength uint16 `struc:"sizeof=Padding"`
	Padding       []byte
}

type respHeader struct {
	Type            byte
	TimeStamp       uint64
	ClientSessionID [8]byte
	PaddingLength   uint16 `struc:"sizeof=Padding"`
	Padding         []byte
}

type cachedUDPState struct {
	sessionAEAD     cipher.AEAD
	sessionRecvAEAD cipher.AEAD
}

func (p *AESUDPClientPacketProcessor) EncodeUDPRequest(request *UDPRequest, out *buf.Buffer,
	cache UDPClientPacketProcessorCachedStateContainer,
) error {
	separateHeaderStruct := separateHeader{PacketID: request.PacketID, SessionID: request.SessionID}
	separateHeaderBuffer := buf.New()
	defer separateHeaderBuffer.Release()
	{
		err := struc.Pack(separateHeaderBuffer, &separateHeaderStruct)
		if err != nil {
			return newError("failed to pack separateHeader").Base(err)
		}
	}
	separateHeaderBufferBytes := separateHeaderBuffer.Bytes()
	{
		encryptedDest := out.Extend(16)
		p.requestSeparateHeaderBlockCipher.Encrypt(encryptedDest, separateHeaderBufferBytes)
	}

	if p.EIHGenerator != nil {
		eih := p.EIHGenerator(separateHeaderBufferBytes[0:16])
		eihHeader := struct {
			EIH ExtensibleIdentityHeaders
		}{
			EIH: eih,
		}
		err := struc.Pack(out, &eihHeader)
		if err != nil {
			return newError("failed to pack eih").Base(err)
		}
	}

	headerStruct := header{
		Type:          UDPHeaderTypeClientToServerStream,
		TimeStamp:     request.TimeStamp,
		PaddingLength: 0,
		Padding:       nil,
	}
	requestBodyBuffer := buf.New()
	{
		err := struc.Pack(requestBodyBuffer, &headerStruct)
		if err != nil {
			return newError("failed to header").Base(err)
		}
	}
	{
		err := addrParser.WriteAddressPort(requestBodyBuffer, request.Address, net.Port(request.Port))
		if err != nil {
			return newError("failed to write address port").Base(err)
		}
	}
	{
		_, err := io.Copy(requestBodyBuffer, bytes.NewReader(request.Payload.Bytes()))
		if err != nil {
			return newError("failed to copy payload").Base(err)
		}
	}
	{
		cacheKey := string(separateHeaderBufferBytes[0:8])
		receivedCacheInterface := cache.GetCachedState(cacheKey)
		cachedState := &cachedUDPState{}
		if receivedCacheInterface != nil {
			cachedState = receivedCacheInterface.(*cachedUDPState)
		}
		if cachedState.sessionAEAD == nil {
			cachedState.sessionAEAD = p.mainPacketAEAD(separateHeaderBufferBytes[0:8])
			cache.PutCachedState(cacheKey, cachedState)
		}

		mainPacketAEADMaterialized := cachedState.sessionAEAD

		encryptedDest := out.Extend(int32(mainPacketAEADMaterialized.Overhead()) + requestBodyBuffer.Len())
		mainPacketAEADMaterialized.Seal(encryptedDest[:0], separateHeaderBuffer.Bytes()[4:16], requestBodyBuffer.Bytes(), nil)
	}
	return nil
}

func (p *AESUDPClientPacketProcessor) DecodeUDPResp(input []byte, resp *UDPResponse,
	cache UDPClientPacketProcessorCachedStateContainer,
) error {
	separateHeaderBuffer := buf.New()
	defer separateHeaderBuffer.Release()
	{
		encryptedDest := separateHeaderBuffer.Extend(16)
		p.responseSeparateHeaderBlockCipher.Decrypt(encryptedDest, input)
	}
	separateHeaderStruct := separateHeader{}
	{
		err := struc.Unpack(separateHeaderBuffer, &separateHeaderStruct)
		if err != nil {
			return newError("failed to unpack separateHeader").Base(err)
		}
	}
	resp.PacketID = separateHeaderStruct.PacketID
	resp.SessionID = separateHeaderStruct.SessionID
	{
		cacheKey := string(separateHeaderBuffer.Bytes()[0:8])
		receivedCacheInterface := cache.GetCachedServerState(cacheKey)
		cachedState := &cachedUDPState{}
		if receivedCacheInterface != nil {
			cachedState = receivedCacheInterface.(*cachedUDPState)
		}

		if cachedState.sessionRecvAEAD == nil {
			cachedState.sessionRecvAEAD = p.mainPacketAEAD(separateHeaderBuffer.Bytes()[0:8])
			cache.PutCachedServerState(cacheKey, cachedState)
		}

		mainPacketAEADMaterialized := cachedState.sessionRecvAEAD
		decryptedDestBuffer := buf.New()
		decryptedDest := decryptedDestBuffer.Extend(int32(len(input)) - 16 - int32(mainPacketAEADMaterialized.Overhead()))
		_, err := mainPacketAEADMaterialized.Open(decryptedDest[:0], separateHeaderBuffer.Bytes()[4:16], input[16:], nil)
		if err != nil {
			return newError("failed to open main packet").Base(err)
		}
		decryptedDestReader := bytes.NewReader(decryptedDest)
		headerStruct := respHeader{}
		{
			err := struc.Unpack(decryptedDestReader, &headerStruct)
			if err != nil {
				return newError("failed to unpack header").Base(err)
			}
		}
		resp.TimeStamp = headerStruct.TimeStamp
		addressReaderBuf := buf.New()
		defer addressReaderBuf.Release()
		var port net.Port
		resp.Address, port, err = addrParser.ReadAddressPort(addressReaderBuf, decryptedDestReader)
		if err != nil {
			return newError("failed to read address port").Base(err)
		}
		resp.Port = int(port)
		readedLength := decryptedDestReader.Size() - int64(decryptedDestReader.Len())
		decryptedDestBuffer.Advance(int32(readedLength))
		resp.Payload = decryptedDestBuffer
		resp.ClientSessionID = headerStruct.ClientSessionID
		return nil
	}
}
