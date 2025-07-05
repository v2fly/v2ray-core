package mirrorenrollment

import (
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorcrypto"
)

type EnrollmentKey struct {
	EnrollmentRequestKey  []byte
	EnrollmentResponseKey []byte
}

func DeriveEnrollmentKeyWithClientAndServerRandom(primaryKey []byte, clientRandom []byte, serverRandom []byte) (*EnrollmentKey, error) {
	requestKey, responseKey, err := mirrorcrypto.DeriveEncryptionKey(primaryKey, clientRandom, serverRandom,
		":connection-enrollment-re78HQNM-CmpRnPbr-PNJVRMhu")
	if err != nil {
		return nil, newError("failed to derive connection enrollment key").Base(err).AtError()
	}
	return &EnrollmentKey{
		EnrollmentRequestKey:  requestKey,
		EnrollmentResponseKey: responseKey,
	}, nil
}

func DeriveEnrollmentServerIdentifier(primaryKey []byte) ([]byte, error) {
	if len(primaryKey) != 32 {
		return nil, newError("invalid primary key size: ", len(primaryKey))
	}

	// Use HKDF to derive a secondary key
	serverID, err := mirrorcrypto.DeriveSecondaryKey(primaryKey, ":connection-enrollment-server-identifier-av38NNGF-TJvRw7C3-p8KM8yKd")
	if err != nil {
		return nil, newError("unable to	server identifier").Base(err)
	}

	return serverID, nil
}
