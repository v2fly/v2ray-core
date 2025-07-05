package mirrorenrollment

import (
	"context"
	"sync"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewServerEnrollmentProcessor(primaryKey []byte) (tlsmirror.ConnectionEnrollmentConfirmationProcessor, error) {
	if len(primaryKey) == 0 {
		return nil, newError("primary key cannot be empty")
	}

	return &serverEnrollmentProcessor{
		primaryKey: primaryKey,
	}, nil
}

type serverEnrollmentProcessor struct {
	primaryKey        []byte
	activeConnections sync.Map
}

func (p *serverEnrollmentProcessor) AddConnection(ctx context.Context, clientRandom, serverRandom []byte, conn tlsmirror.InsertableTLSConnForEnrollment) (tlsmirror.RemoveConnectionFunc, error) {
	if conn == nil {
		return nil, newError("nil InsertableTLSConnForEnrollment")
	}

	enrollmentKey, err := DeriveEnrollmentKeyWithClientAndServerRandom(p.primaryKey, clientRandom, serverRandom)
	if err != nil {
		return nil, newError("failed to derive enrollment key").Base(err).AtError()
	}

	if _, loaded := p.activeConnections.LoadOrStore(string(enrollmentKey.EnrollmentRequestKey), conn); loaded {
		return nil, newError("connection with ID ", enrollmentKey.EnrollmentRequestKey, " already exists")
	}

	return func() error {
		p.activeConnections.Delete(string(enrollmentKey.EnrollmentRequestKey))
		return nil
	}, nil
}

func (p *serverEnrollmentProcessor) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	if req == nil {
		return nil, newError("nil EnrollmentConfirmationReq")
	}

	if req.CarrierTlsConnectionClientRandom == nil || req.CarrierTlsConnectionServerRandom == nil {
		return nil, newError("missing client or server random in EnrollmentConfirmationReq")
	}

	enrollmentKey, err := DeriveEnrollmentKeyWithClientAndServerRandom(p.primaryKey,
		req.CarrierTlsConnectionClientRandom, req.CarrierTlsConnectionServerRandom)
	if err != nil {
		return nil, newError("failed to derive enrollment key").Base(err).AtError()
	}

	if conn, ok := p.activeConnections.Load(string(enrollmentKey.EnrollmentRequestKey)); ok {
		insertableConn, ok := conn.(tlsmirror.InsertableTLSConnForEnrollment)
		if !ok {
			return nil, newError("connection does not implement InsertableTLSConnForEnrollment")
		}

		return insertableConn.VerifyConnectionEnrollment(req)
	}

	return &tlsmirror.EnrollmentConfirmationResp{
		Enrolled: false,
	}, nil
}
