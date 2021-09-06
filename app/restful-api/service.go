package restful_api

import (
	"context"
	"github.com/gin-gonic/gin"
	"net"
	"sync"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type restfulService struct {
	*gin.Engine

	listener net.Listener
	config   *Config
	access   sync.Mutex

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
