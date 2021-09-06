package restful_api

import (
	"context"
	"github.com/gin-gonic/gin"
	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/features"
	feature_stats "github.com/v2fly/v2ray-core/v4/features/stats"
	"net"
	"sync"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type restfulService struct {
	*gin.Engine

	listener net.Listener
	config   *Config
	access   sync.Mutex

	stats feature_stats.Manager

	ctx context.Context
}

func (r *restfulService) Type() interface{} {
	return (*struct{})(nil)
}

func (r *restfulService) Start() error {
	defer r.access.Unlock()
	r.access.Lock()
	return r.start()
}

func (r *restfulService) Close() error {
	defer r.access.Unlock()
	r.access.Lock()
	if r.listener != nil {
		return r.listener.Close()
	}
	return nil
}

func (r *restfulService) init(config *Config, stats feature_stats.Manager) {
	r.stats = stats
	r.config = config
}

func newRestfulService(ctx context.Context, config *Config) (features.Feature, error) {
	r := new(restfulService)
	r.ctx = ctx
	if err := core.RequireFeatures(ctx, func(stats feature_stats.Manager) {
		r.init(config, stats)
	}); err != nil {
		return nil, err
	}
	return r, nil
}
