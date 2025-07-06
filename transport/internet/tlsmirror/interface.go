package tlsmirror

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
)

type TLSRecord struct {
	RecordType            byte
	LegacyProtocolVersion [2]byte
	RecordLength          uint16
	Fragment              []byte

	// Annotations are used to store additional information about the record. Never sent over the wire.
	InsertedMessage bool
}

type RecordReader interface {
	ReadNextRecord(rejectProfile PartialTLSRecordRejectProfile) (*TLSRecord, error)
}

type RecordWriter interface {
	WriteRecord(record *TLSRecord) error
}

type Peeker interface {
	Peek(n int) ([]byte, error)
}

type PartialTLSRecordRejectProfile interface {
	TestIfReject(record *TLSRecord, readyFields int) error
}

type MessageHook func(message *TLSRecord) (drop bool, ok error)

type ExplicitNonceDetection func(cipherSuite uint16) bool

type InsertableTLSConn interface {
	common.Closable
	GetHandshakeRandom() ([]byte, []byte, error)
	InsertC2SMessage(message *TLSRecord) error
	InsertS2CMessage(message *TLSRecord) error
	GetApplicationDataExplicitNonceReservedOverheadHeaderLength() (int, error)
}

const TrafficGeneratorManagedConnectionContextKey = "TrafficGeneratorManagedConnection-ku63HMMD-kduCPhr8-DN4y6WEa"

type TrafficGeneratorManagedConnection interface {
	RecallTrafficGenerator() error
	WaitConnectionReady() context.Context
	IsConnectionInvalidated() bool
}

type ConnectionEnrollmentConfirmation interface {
	VerifyConnectionEnrollment(req *EnrollmentConfirmationReq) (*EnrollmentConfirmationResp, error)
}

const EnrollmentVerificationControlConnectionPostfix = ".tlsmirror-controlconnection.v2fly.arpa"

type InsertableTLSConnForEnrollment interface {
	InsertableTLSConnEnrollmentEventReceiver
}

type InsertableTLSConnEnrollmentEventReceiver interface {
	ConnectionEnrollmentConfirmation
}

type RemoveConnectionFunc func() error

type ConnectionEnrollmentConfirmationProcessor interface {
	ConnectionEnrollmentConfirmation
	AddConnection(ctx context.Context, clientRandom, ServerRandom []byte, conn InsertableTLSConnForEnrollment) (RemoveConnectionFunc, error)
}
