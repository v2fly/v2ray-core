package httpenrollmentconfirmation

import (
	"io"
	"net/http"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
)

func NewHTTPEnrollmentConfirmationServerFromConnectionEnrollmentConfirmation(confirmation tlsmirror.ConnectionEnrollmentConfirmation) (http.Handler, error) {
	if confirmation == nil {
		return nil, newError("nil confirmation")
	}
	return &server{
		ConnectionEnrollmentConfirmation: confirmation,
	}, nil
}

type server struct {
	tlsmirror.ConnectionEnrollmentConfirmation
}

func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "failed to read request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req := &tlsmirror.EnrollmentConfirmationReq{}
	if err := proto.Unmarshal(requestBody, req); err != nil {
		http.Error(writer, "failed to unmarshal request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.VerifyConnectionEnrollment(req)
	if err != nil {
		http.Error(writer, "failed to verify connection enrollment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	responseBody, err := proto.Marshal(resp)
	if err != nil {
		http.Error(writer, "failed to marshal response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(responseBody); err != nil {
		http.Error(writer, "failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		http.Error(writer, "response writer does not support flushing", http.StatusInternalServerError)
		return
	}
}
