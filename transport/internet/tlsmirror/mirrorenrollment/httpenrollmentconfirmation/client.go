package httpenrollmentconfirmation

import (
	"bytes"
	"encoding/base32"
	"io"
	"net/http"

	protov2 "google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewHTTPEnrollmentConfirmationClientFromHTTPRoundTripper(tripper http.RoundTripper) (tlsmirror.ConnectionEnrollmentConfirmation, error) {
	if tripper == nil {
		return nil, newError("nil tripper")
	}
	return &client{
		httpRoundTripper: tripper,
	}, nil
}

type client struct {
	httpRoundTripper http.RoundTripper
}

func (c *client) VerifyConnectionEnrollment(req *tlsmirror.EnrollmentConfirmationReq) (*tlsmirror.EnrollmentConfirmationResp, error) {
	requestMessage, err := protov2.Marshal(req)
	if err != nil {
		return nil, newError("failed to marshal enrollment confirmation request").Base(err)
	}

	serverID := base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding).EncodeToString(req.ServerIdentifier)

	httpReq, err := http.NewRequest("POST", "http://"+serverID+tlsmirror.EnrollmentVerificationControlConnectionPostfix, bytes.NewReader(requestMessage))
	if err != nil {
		return nil, newError("failed to create HTTP request").Base(err)
	}
	httpResp, err := c.httpRoundTripper.RoundTrip(httpReq)
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()
	if err != nil {
		return nil, newError("failed to send HTTP request").Base(err)
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, newError("unexpected HTTP response status: ", httpResp.StatusCode)
	}
	responseMessage, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, newError("failed to read HTTP response body").Base(err)
	}
	resp := &tlsmirror.EnrollmentConfirmationResp{}
	if err := protov2.Unmarshal(responseMessage, resp); err != nil {
		return nil, newError("failed to unmarshal enrollment confirmation response").Base(err)
	}
	return resp, nil
}
