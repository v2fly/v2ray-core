package command

//go:generate go run v2ray.com/core/common/errors/errorgen

import (
	"context"
	"time"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/features/routing"
	"v2ray.com/core/features/stats"
)

// routingServer is an implementation of RoutingService.
type routingServer struct {
	router       routing.Router
	routingStats stats.Channel
}

// NewRoutingServer creates a statistics service with statistics manager.
func NewRoutingServer(router routing.Router, routingStats stats.Channel) RoutingServiceServer {
	return &routingServer{
		router:       router,
		routingStats: routingStats,
	}
}

func (s *routingServer) TestRoute(ctx context.Context, request *TestRouteRequest) (*RoutingContext, error) {
	if request.RoutingContext == nil {
		return nil, newError("Invalid routing request.")
	}
	route, err := s.router.PickRoute(AsRoutingContext(request.RoutingContext))
	if err != nil {
		return nil, err
	}
	if request.PublishResult && s.routingStats != nil {
		ctx, _ := context.WithTimeout(context.Background(), 4*time.Second) // nolint: govet
		s.routingStats.Publish(ctx, route)
	}
	return AsProtobufMessage(request.FieldSelectors)(route), nil
}

func (s *routingServer) SubscribeRoutingStats(request *SubscribeRoutingStatsRequest, stream RoutingService_SubscribeRoutingStatsServer) error {
	if s.routingStats == nil {
		return newError("Routing statistics not enabled.")
	}
	genMessage := AsProtobufMessage(request.FieldSelectors)
	subscriber, err := stats.SubscribeRunnableChannel(s.routingStats)
	if err != nil {
		return err
	}
	defer stats.UnsubscribeClosableChannel(s.routingStats, subscriber)
	for {
		select {
		case value, ok := <-subscriber:
			if !ok {
				return newError("Upstream closed the subscriber channel.")
			}
			route, ok := value.(routing.Route)
			if !ok {
				return newError("Upstream sent malformed statistics.")
			}
			err := stream.Send(genMessage(route))
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func (s *routingServer) GetHealthInfo(ctx context.Context, request *GetHealthInfoRequest) (*GetHealthInfoResponse, error) {
	h, ok := s.router.(routing.RouterChecker)
	if !ok {
		return nil, status.Errorf(codes.Unavailable, "current router is not a health checker")
	}
	results, err := h.GetBalancersInfo(request.BalancerTags)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	rsp := &GetHealthInfoResponse{
		Balancers: make([]*BalancerInfo, 0),
	}
	for _, result := range results {
		stat := &BalancerInfo{
			Tag:      result.Tag,
			Strategy: result.Strategy.Name,
			Titles:   result.Strategy.ValueTitles,
			Selects:  make([]*OutboundInfo, 0),
			Others:   make([]*OutboundInfo, 0),
		}
		for _, item := range result.Strategy.Selects {
			stat.Selects = append(stat.Selects, &OutboundInfo{
				Tag:    item.Tag,
				Values: item.Values,
			})
		}
		for _, item := range result.Strategy.Others {
			stat.Others = append(stat.Others, &OutboundInfo{
				Tag:    item.Tag,
				Values: item.Values,
			})
		}
		rsp.Balancers = append(rsp.Balancers, stat)
	}
	return rsp, nil
}
func (s *routingServer) CheckBalancers(ctx context.Context, request *CheckBalancersRequest) (*CheckBalancersResponse, error) {
	h, ok := s.router.(routing.RouterChecker)
	if !ok {
		return nil, status.Errorf(codes.Unavailable, "current router is not a health checker")
	}
	go func() {
		err := h.CheckBalancers(request.BalancerTags)
		if err != nil {
			newError("CheckBalancers error:", err).AtInfo().WriteToLog()
		}
	}()
	return &CheckBalancersResponse{}, nil
}

func (s *routingServer) mustEmbedUnimplementedRoutingServiceServer() {}

type service struct {
	v *core.Instance
}

func (s *service) Register(server *grpc.Server) {
	common.Must(s.v.RequireFeatures(func(router routing.Router, stats stats.Manager) {
		RegisterRoutingServiceServer(server, NewRoutingServer(router, nil))
	}))
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		s := core.MustFromContext(ctx)
		return &service{v: s}, nil
	}))
}
