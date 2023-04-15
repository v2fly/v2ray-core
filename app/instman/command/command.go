package command

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

type service struct {
	UnimplementedInstanceManagementServiceServer

	instman extension.InstanceManagement
}

func (s service) ListInstance(ctx context.Context, req *ListInstanceReq) (*ListInstanceResp, error) {
	instanceNames, err := s.instman.ListInstance(ctx)
	if err != nil {
		return nil, err
	}
	return &ListInstanceResp{Name: instanceNames}, nil
}

func (s service) AddInstance(ctx context.Context, req *AddInstanceReq) (*AddInstanceResp, error) {
	configContent, err := base64.StdEncoding.DecodeString(req.ConfigContentB64)
	if err != nil {
		return nil, err
	}
	err = s.instman.AddInstance(ctx, req.Name, configContent, req.ConfigType)
	if err != nil {
		return nil, err
	}
	return &AddInstanceResp{}, nil
}

func (s service) StartInstance(ctx context.Context, req *StartInstanceReq) (*StartInstanceResp, error) {
	err := s.instman.StartInstance(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &StartInstanceResp{}, nil
}

func (s service) Register(server *grpc.Server) {
	RegisterInstanceManagementServiceServer(server, s)
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		s := core.MustFromContext(ctx)
		sv := &service{}
		err := s.RequireFeatures(func(instman extension.InstanceManagement) {
			sv.instman = instman
		})
		if err != nil {
			return nil, err
		}
		return sv, nil
	}))
}
