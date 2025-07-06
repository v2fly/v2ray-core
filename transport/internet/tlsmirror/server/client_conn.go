package server

import (
	"bytes"
	"context"
	"crypto/cipher"
	gonet "net"
	"time"

	"golang.org/x/crypto/chacha20"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcrypto"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment"
)

type clientConnState struct {
	ctx     context.Context
	done    context.CancelFunc
	handler internet.ConnHandler

	mirrorConn tlsmirror.InsertableTLSConn
	localAddr  net.Addr
	remoteAddr net.Addr

	activated bool
	decryptor *mirrorcrypto.Decryptor
	encryptor *mirrorcrypto.Encryptor

	primaryKey []byte

	readPipe   chan []byte
	readBuffer *bytes.Buffer

	protocolVersion [2]byte

	firstWrite      bool
	firstWriteDelay time.Duration

	transportLayerPadding *TransportLayerPadding

	sequenceWatermarkEnabled                 bool
	sequenceWatermarkTx, sequenceWatermarkRx cipher.Stream
}

func (s *clientConnState) GetConnectionContext() context.Context {
	return s.ctx
}

func (s *clientConnState) Read(b []byte) (n int, err error) {
	if s.readBuffer != nil {
		n, _ = s.readBuffer.Read(b)
		if n > 0 {
			return n, nil
		}
		s.readBuffer = nil
	}

	select {
	case <-s.ctx.Done():
		return 0, s.ctx.Err()
	case data := <-s.readPipe:
		if s.transportLayerPadding != nil && s.transportLayerPadding.Enabled {
			var padding int
			data, padding = Unpack(data)
			_ = padding
			if data == nil {
				return 0, nil
			}
		}
		s.readBuffer = bytes.NewBuffer(data)
		n, err = s.readBuffer.Read(b)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
}

func (s *clientConnState) Write(b []byte) (n int, err error) {
	payloadSize := len(b)
	if s.firstWrite && s.firstWriteDelay > 0 {
		firstWriteDelayTimer := time.NewTimer(s.firstWriteDelay)
		defer firstWriteDelayTimer.Stop()
		select {
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		case <-firstWriteDelayTimer.C:
			s.firstWrite = false
		}
	}
	if s.transportLayerPadding != nil && s.transportLayerPadding.Enabled {
		b = Pack(bytes.Clone(b), 0)
	}
	err = s.WriteMessage(b)
	if err != nil {
		return 0, err
	}
	n = payloadSize
	return n, nil
}

func (s *clientConnState) Close() error {
	// TODO: Only recall the connection, not close it.
	s.done()
	return nil
}

func (s *clientConnState) LocalAddr() gonet.Addr {
	return s.remoteAddr
}

func (s *clientConnState) RemoteAddr() gonet.Addr {
	return s.remoteAddr
}

func (s *clientConnState) SetDeadline(t time.Time) error {
	return nil
}

func (s *clientConnState) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *clientConnState) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s *clientConnState) onC2SMessage(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	if message.RecordType == mirrorcommon.TLSRecord_RecordType_application_data {
		if s.decryptor == nil {
			clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
			if err != nil {
				newError("failed to get handshake random").Base(err).AtWarning().WriteToLog()
				return false, nil
			}

			{
				encryptionKey, nonceMask, err := mirrorcrypto.DeriveEncryptionKey(s.primaryKey, clientRandom, serverRandom, ":s2c")
				if err != nil {
					newError("failed to derive C2S encryption key").Base(err).AtWarning().WriteToLog()
					return false, nil
				}
				s.decryptor = mirrorcrypto.NewDecryptor(encryptionKey, nonceMask)
			}

			{
				encryptionKey, nonceMask, err := mirrorcrypto.DeriveEncryptionKey(s.primaryKey, clientRandom, serverRandom, ":c2s")
				if err != nil {
					newError("failed to derive S2C encryption key").Base(err).AtWarning().WriteToLog()
					return false, nil
				}
				s.encryptor = mirrorcrypto.NewEncryptor(encryptionKey, nonceMask)
			}
			s.protocolVersion = message.LegacyProtocolVersion

			if !s.activated {
				s.handler(s)
				s.activated = true
			}
		}
	}
	return false, ok
}

func (s *clientConnState) onS2CMessage(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	if s.sequenceWatermarkEnabled {
		if s.sequenceWatermarkRx != nil {
			if (message.RecordType == mirrorcommon.TLSRecord_RecordType_application_data ||
				message.RecordType == mirrorcommon.TLSRecord_RecordType_alert) && len(message.Fragment) >= 16 {
				watermarkRegion := message.Fragment[len(message.Fragment)-16:]
				s.sequenceWatermarkRx.XORKeyStream(watermarkRegion, watermarkRegion)
			}
		}
	}

	if message.RecordType == mirrorcommon.TLSRecord_RecordType_application_data {
		explicitNonceReservedOverheadHeaderLength, err := s.mirrorConn.GetApplicationDataExplicitNonceReservedOverheadHeaderLength()
		if err != nil {
			return false, newError("failed to get explicit nonce reserved overhead header length").Base(err)
		}

		if s.encryptor == nil {
			return false, nil
		}
		buffer := make([]byte, 0, len(message.Fragment)-s.encryptor.NonceSize()-explicitNonceReservedOverheadHeaderLength)
		buffer, err = s.decryptor.Open(buffer, message.Fragment[explicitNonceReservedOverheadHeaderLength:])
		if err != nil {
			return false, nil
		}

		if s.sequenceWatermarkEnabled && s.sequenceWatermarkRx == nil {
			clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
			if err != nil {
				newError("failed to get handshake random").Base(err).AtError().WriteToLog()
				return true, nil
			}
			key, nonce, err := mirrorcrypto.DeriveSequenceWatermarkingKey(s.primaryKey, clientRandom, serverRandom, ":s2c")
			if err != nil {
				newError("failed to derive sequence watermarking key").Base(err).AtError().WriteToLog()
				return true, nil
			}
			s.sequenceWatermarkRx, err = chacha20.NewUnauthenticatedCipher(key, nonce)
			if err != nil {
				newError("failed to create sequence watermarking cipher").Base(err).AtError().WriteToLog()
				return true, nil
			}
		}

		s.readPipe <- buffer
		return true, nil
	}
	return false, ok
}

func (s *clientConnState) WriteMessage(message []byte) error {
	explicitNonceReservedOverheadHeaderLength, err := s.mirrorConn.GetApplicationDataExplicitNonceReservedOverheadHeaderLength()
	if err != nil {
		return newError("failed to get explicit nonce reserved overhead header length").Base(err)
	}
	buffer := make([]byte, explicitNonceReservedOverheadHeaderLength, explicitNonceReservedOverheadHeaderLength+len(message)+s.encryptor.NonceSize())
	buffer, err = s.encryptor.Seal(buffer, message)
	if err != nil {
		return newError("failed to encrypt message").Base(err)
	}
	record := tlsmirror.TLSRecord{
		RecordType:            mirrorcommon.TLSRecord_RecordType_application_data,
		LegacyProtocolVersion: s.protocolVersion,
		RecordLength:          uint16(len(buffer)),
		Fragment:              buffer,
		InsertedMessage:       true,
	}
	return s.mirrorConn.InsertC2SMessage(&record)
}

func (s *clientConnState) VerifyConnectionEnrollmentWithProcessor(connectionEnrollmentConfirmationClient tlsmirror.ConnectionEnrollmentConfirmation) error {
	clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
	if err != nil {
		return newError("failed to get handshake random").Base(err).AtWarning()
	}

	enrollmentServerIdentity, err := mirrorenrollment.DeriveEnrollmentServerIdentifier(s.primaryKey)
	if err != nil {
		return newError("failed to derive enrollment server identifier").Base(err).AtError()
	}

	enrollmentReq := &tlsmirror.EnrollmentConfirmationReq{
		ServerIdentifier:                 enrollmentServerIdentity,
		CarrierTlsConnectionClientRandom: clientRandom,
		CarrierTlsConnectionServerRandom: serverRandom,
	}
	resp, err := connectionEnrollmentConfirmationClient.VerifyConnectionEnrollment(enrollmentReq)
	if err != nil {
		return newError("failed to verify connection enrollment").Base(err).AtError()
	}
	if resp == nil {
		return newError("nil EnrollmentConfirmationResp")
	}
	if resp.Enrolled {
		newError("connection enrolled successfully").AtInfo().WriteToLog()
	} else {
		return newError("connection enrollment failed")
	}
	return nil
}

func (s *clientConnState) onC2SMessageTx(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	if s.sequenceWatermarkEnabled {
		if s.sequenceWatermarkTx != nil {
			if (message.RecordType == mirrorcommon.TLSRecord_RecordType_application_data ||
				message.RecordType == mirrorcommon.TLSRecord_RecordType_alert) && len(message.Fragment) >= 16 {
				watermarkRegion := message.Fragment[len(message.Fragment)-16:]
				s.sequenceWatermarkTx.XORKeyStream(watermarkRegion, watermarkRegion)
			}
		}
		if message.InsertedMessage && s.sequenceWatermarkTx == nil {
			clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
			if err != nil {
				newError("failed to get handshake random").Base(err).AtError().WriteToLog()
				return true, nil
			}
			key, nonce, err := mirrorcrypto.DeriveSequenceWatermarkingKey(s.primaryKey, clientRandom, serverRandom, ":c2s")
			if err != nil {
				newError("failed to derive sequence watermarking key").Base(err).AtError().WriteToLog()
				return true, nil
			}
			s.sequenceWatermarkTx, err = chacha20.NewUnauthenticatedCipher(key, nonce)
			if err != nil {
				newError("failed to create sequence watermarking cipher").Base(err).AtError().WriteToLog()
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *clientConnState) onS2CMessageTx(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	return false, nil
}
