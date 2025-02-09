package command

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/subscription"
	"github.com/v2fly/v2ray-core/v5/common"

	"google.golang.org/grpc"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type SubscriptionManagerService struct {
	UnimplementedSubscriptionManagerServiceServer
	manager subscription.SubscriptionManager
}

func (s *SubscriptionManagerService) UpdateTrackedSubscription(ctx context.Context, request *UpdateTrackedSubscriptionRequest) (*UpdateTrackedSubscriptionResponse, error) {
	if s.manager == nil {
		return nil, newError("subscription manager is not available")
	}
	err := s.manager.UpdateTrackedSubscription(request.Name)
	if err != nil {
		return nil, err
	}
	return &UpdateTrackedSubscriptionResponse{}, nil
}

func NewSubscriptionManagerService(manager subscription.SubscriptionManager) *SubscriptionManagerService {
	return &SubscriptionManagerService{manager: manager}
}

func (s *SubscriptionManagerService) ListTrackedSubscription(ctx context.Context, req *ListTrackedSubscriptionRequest) (*ListTrackedSubscriptionResponse, error) {
	if s.manager == nil {
		return nil, newError("subscription manager is not available")
	}
	names := s.manager.ListTrackedSubscriptions()
	return &ListTrackedSubscriptionResponse{Names: names}, nil
}

func (s *SubscriptionManagerService) AddTrackedSubscription(ctx context.Context, req *AddTrackedSubscriptionRequest) (*AddTrackedSubscriptionResponse, error) {
	if s.manager == nil {
		return nil, newError("subscription manager is not available")
	}
	err := s.manager.AddTrackedSubscriptionFromImportSource(req.Source)
	if err != nil {
		return nil, err
	}
	return &AddTrackedSubscriptionResponse{}, nil
}

func (s *SubscriptionManagerService) RemoveTrackedSubscription(ctx context.Context, req *RemoveTrackedSubscriptionRequest) (*RemoveTrackedSubscriptionResponse, error) {
	if s.manager == nil {
		return nil, newError("subscription manager is not available")
	}
	err := s.manager.RemoveTrackedSubscription(req.Name)
	if err != nil {
		return nil, err
	}
	return &RemoveTrackedSubscriptionResponse{}, nil
}

func (s *SubscriptionManagerService) GetTrackedSubscriptionStatus(ctx context.Context, req *GetTrackedSubscriptionStatusRequest) (*GetTrackedSubscriptionStatusResponse, error) {
	if s.manager == nil {
		return nil, newError("subscription manager is not available")
	}
	status, err := s.manager.GetTrackedSubscriptionStatus(req.Name)
	if err != nil {
		return nil, err
	}
	return &GetTrackedSubscriptionStatusResponse{Status: status}, nil
}

func (s *SubscriptionManagerService) Register(server *grpc.Server) {
	RegisterSubscriptionManagerServiceServer(server, s)
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		var manager subscription.SubscriptionManager
		common.Must(core.RequireFeatures(ctx, func(m subscription.SubscriptionManager) {
			manager = m
		}))
		service := NewSubscriptionManagerService(manager)
		return service, nil
	}))
}
