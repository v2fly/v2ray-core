package request

import "context"

type ReverserImpl interface {
	OnOtherRoundTrip(ctx context.Context, req Request, opts ...RoundTripperOption) (resp Response, err error)
	OnAuthenticatedServerIntentRoundTrip(ctx context.Context, serverPublic []byte, req Request, opts ...RoundTripperOption) (resp Response, err error)
}

type ReverserAccessChecker interface {
	CheckReverserAccess(ctx context.Context, serverKey []byte) (clientKey []byte, err error)
}
