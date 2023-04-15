package command

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"

	grpc "google.golang.org/grpc"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/common"
	cmlog "github.com/v2fly/v2ray-core/v5/common/log"
)

// LoggerServer is the implemention of LoggerService
type LoggerServer struct {
	V *core.Instance
}

// RestartLogger implements LoggerService.
func (s *LoggerServer) RestartLogger(ctx context.Context, request *RestartLoggerRequest) (*RestartLoggerResponse, error) {
	logger := s.V.GetFeature((*log.Instance)(nil))
	if logger == nil {
		return nil, newError("unable to get logger instance")
	}
	if err := logger.Close(); err != nil {
		return nil, newError("failed to close logger").Base(err)
	}
	if err := logger.Start(); err != nil {
		return nil, newError("failed to start logger").Base(err)
	}
	return &RestartLoggerResponse{}, nil
}

// FollowLog implements LoggerService.
func (s *LoggerServer) FollowLog(_ *FollowLogRequest, stream LoggerService_FollowLogServer) error {
	logger := s.V.GetFeature((*log.Instance)(nil))
	if logger == nil {
		return newError("unable to get logger instance")
	}
	follower, ok := logger.(cmlog.Follower)
	if !ok {
		return newError("logger not support following")
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	f := func(msg cmlog.Message) {
		err := stream.Send(&FollowLogResponse{
			Message: msg.String(),
		})
		if err != nil {
			cancel()
		}
	}
	follower.AddFollower(f)
	defer follower.RemoveFollower(f)
	<-ctx.Done()
	return nil
}

func (s *LoggerServer) mustEmbedUnimplementedLoggerServiceServer() {}

type service struct {
	v *core.Instance
}

func (s *service) Register(server *grpc.Server) {
	RegisterLoggerServiceServer(server, &LoggerServer{
		V: s.v,
	})
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		s := core.MustFromContext(ctx)
		return &service{v: s}, nil
	}))
}
