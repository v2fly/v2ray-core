package webcommander

import (
	"archive/zip"
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/commander"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func newWebCommander(ctx context.Context, config *Config) (*WebCommander, error) {
	if config == nil {
		return nil, newError("config is nil")
	}
	if config.Tag == "" {
		return nil, newError("config.Tag is empty")
	}
	var webRootfs fs.FS
	if config.WebRoot != nil {
		zipReader, err := zip.NewReader(bytes.NewReader(config.WebRoot), int64(len(config.WebRoot)))
		if err != nil {
			return nil, newError("failed to create zip reader").Base(err)
		}
		webRootfs = zipReader
	}

	return &WebCommander{ctx: ctx, config: config, webRootfs: webRootfs}, nil
}

type WebCommander struct {
	sync.Mutex

	ctx context.Context
	ohm outbound.Manager
	cm  commander.CommanderIfce

	server      *http.Server
	wrappedGrpc *grpcweb.WrappedGrpcServer
	webRootfs   fs.FS

	config *Config
}

func (w *WebCommander) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	apiPath := w.config.ApiMountpoint
	if strings.HasPrefix(request.URL.Path, apiPath) {
		request.URL.Path = strings.TrimPrefix(request.URL.Path, apiPath)
		if w.wrappedGrpc.IsGrpcWebRequest(request) {
			w.wrappedGrpc.ServeHTTP(writer, request)
			return
		}
	}
	if w.webRootfs != nil {
		http.ServeFileFS(writer, request, w.webRootfs, request.URL.Path)
		return
	}
	writer.WriteHeader(http.StatusNotFound)
}

func (w *WebCommander) asyncStart() {
	var grpcServer *grpc.Server
	for {
		grpcServer = w.cm.ExtractGrpcServer()
		if grpcServer != nil {
			break
		}
		time.Sleep(time.Second)
	}

	listener := commander.NewOutboundListener()

	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	w.server = &http.Server{}
	w.wrappedGrpc = wrappedGrpc
	w.server.Handler = w

	go func() {
		err := w.server.Serve(listener)
		if err != nil {
			newError("failed to serve HTTP").Base(err).WriteToLog()
		}
	}()

	if err := w.ohm.RemoveHandler(context.Background(), w.config.Tag); err != nil {
		newError("failed to remove existing handler").WriteToLog()
	}

	if err := w.ohm.AddHandler(context.Background(), commander.NewOutbound(w.config.Tag, listener)); err != nil {
		newError("failed to add handler").Base(err).WriteToLog()
	}
}

func (w *WebCommander) Type() interface{} {
	return (*WebCommander)(nil)
}

func (w *WebCommander) Start() error {
	if err := core.RequireFeatures(w.ctx, func(cm commander.CommanderIfce, om outbound.Manager) {
		w.Lock()
		defer w.Unlock()

		w.cm = cm
		w.ohm = om

		go w.asyncStart()
	}); err != nil {
		return err
	}

	return nil
}

func (w *WebCommander) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.server != nil {
		if err := w.server.Close(); err != nil {
			return newError("failed to close http server").Base(err)
		}

		w.server = nil
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return newWebCommander(ctx, config.(*Config))
	}))
}
