package server

import (
	"bytes"
	"context"
	"crypto/cipher"
	"net"
	"time"

	"golang.org/x/crypto/chacha20"

	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcommon"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcrypto"
)

type connState struct {
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

	connectionEnrollmentEnabled           bool
	connectionEnrollmentProcessorEnrolled bool
	connectionEnrollmentProcessor         tlsmirror.ConnectionEnrollmentConfirmationProcessor
	connectionEnrollmentRemover           tlsmirror.RemoveConnectionFunc

	sequenceWatermarkEnabled                 bool
	sequenceWatermarkTx, sequenceWatermarkRx cipher.Stream
}

func (s *connState) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	return &tlsmirror.EnrollmentConfirmationResp{Enrolled: true}, nil
}

func (s *connState) Read(b []byte) (n int, err error) {
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

func (s *connState) Write(b []byte) (n int, err error) {
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

func (s *connState) LocalAddr() net.Addr {
	return s.localAddr
}

func (s *connState) RemoteAddr() net.Addr {
	return s.remoteAddr
}

func (s *connState) SetDeadline(t time.Time) error {
	return nil
}

func (s *connState) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *connState) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s *connState) Close() error {
	s.done()
	if s.connectionEnrollmentRemover != nil {
		if err := s.connectionEnrollmentRemover(); err != nil {
			newError("failed to remove connection enrollment").Base(err).AtWarning().WriteToLog()
		}
		s.connectionEnrollmentRemover = nil
	}
	return nil
}

func (s *connState) onS2CMessage(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	if message.RecordType == mirrorcommon.TLSRecord_RecordType_handshake {
		if s.connectionEnrollmentEnabled && !s.connectionEnrollmentProcessorEnrolled {
			clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
			if err != nil {
				newError("failed to get handshake random").Base(err).AtWarning().WriteToLog()
				return false, nil
			}

			remove, err := s.connectionEnrollmentProcessor.AddConnection(s.ctx, clientRandom, serverRandom, s)
			if err != nil {
				newError("failed to add connection for enrollment").Base(err).AtWarning().WriteToLog()
				return false, nil
			}
			s.connectionEnrollmentRemover = remove
			s.connectionEnrollmentProcessorEnrolled = true
			return false, nil
		}
	}
	return false, nil
}

func (s *connState) onC2SMessage(message *tlsmirror.TLSRecord) (drop bool, ok error) {
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
		if s.decryptor == nil {
			clientRandom, serverRandom, err := s.mirrorConn.GetHandshakeRandom()
			if err != nil {
				newError("failed to get handshake random").Base(err).AtWarning().WriteToLog()
				return false, nil
			}

			{
				encryptionKey, nonceMask, err := mirrorcrypto.DeriveEncryptionKey(s.primaryKey, clientRandom, serverRandom, ":c2s")
				if err != nil {
					newError("failed to derive C2S encryption key").Base(err).AtWarning().WriteToLog()
					return false, nil
				}
				s.decryptor = mirrorcrypto.NewDecryptor(encryptionKey, nonceMask)
			}

			{
				encryptionKey, nonceMask, err := mirrorcrypto.DeriveEncryptionKey(s.primaryKey, clientRandom, serverRandom, ":s2c")
				if err != nil {
					newError("failed to derive S2C encryption key").Base(err).AtWarning().WriteToLog()
					return false, nil
				}
				s.encryptor = mirrorcrypto.NewEncryptor(encryptionKey, nonceMask)
			}
			s.protocolVersion = message.LegacyProtocolVersion
		}

		explicitNonceReservedOverheadHeaderLength, err := s.mirrorConn.GetApplicationDataExplicitNonceReservedOverheadHeaderLength()
		if err != nil {
			return false, newError("failed to get explicit nonce reserved overhead header length").Base(err)
		}

		buffer := make([]byte, 0, len(message.Fragment)-s.decryptor.NonceSize()-explicitNonceReservedOverheadHeaderLength)
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
			key, nonce, err := mirrorcrypto.DeriveSequenceWatermarkingKey(s.primaryKey, clientRandom, serverRandom, ":c2s")
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

		if !s.activated {
			s.handler(s)
			s.activated = true
		}
		s.readPipe <- buffer
		return true, nil
	}
	return false, ok
}

func (s *connState) WriteMessage(message []byte) error {
	explicitNonceReservedOverheadHeaderLength, err := s.mirrorConn.GetApplicationDataExplicitNonceReservedOverheadHeaderLength()
	if err != nil {
		return newError("failed to get explicit nonce reserved overhead header length").Base(err)
	}

	buffer := make([]byte, explicitNonceReservedOverheadHeaderLength, explicitNonceReservedOverheadHeaderLength+len(message)+s.decryptor.NonceSize())
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
	return s.mirrorConn.InsertS2CMessage(&record)
}

func (s *connState) onC2SMessageTx(message *tlsmirror.TLSRecord) (drop bool, ok error) {
	return false, nil
}

func (s *connState) onS2CMessageTx(message *tlsmirror.TLSRecord) (drop bool, ok error) {
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
			key, nonce, err := mirrorcrypto.DeriveSequenceWatermarkingKey(s.primaryKey, clientRandom, serverRandom, ":s2c")
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
