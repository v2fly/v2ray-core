// +build !confonly

package browserforwarder

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/v2fly/BrowserBridge/handler"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform/securedload"
	"github.com/v2fly/v2ray-core/v4/features/ext"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type Forwarder struct {
	ctx context.Context

	forwarder  *handler.HTTPHandle
	httpserver *http.Server

	config *Config
}

func (f *Forwarder) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	requestPath := request.URL.Path[1:]

	switch requestPath {
	case "":
		fallthrough
	case "index.js":
		BridgeResource(writer, request, requestPath)
	case "link":
		f.forwarder.ServeBridge(writer, request)
	}
}

func (f *Forwarder) DialWebsocket(url string, header http.Header) (io.ReadWriteCloser, error) {
	return f.forwarder.Dial(url)
}

func (f *Forwarder) Type() interface{} {
	return ext.BrowserForwarderType()
}

func (f *Forwarder) Start() error {
	f.forwarder = handler.NewHttpHandle()
	f.httpserver = &http.Server{Handler: f}
	address := net.ParseAddress(f.config.ListenAddr)
	listener, err := internet.ListenSystem(f.ctx, &net.TCPAddr{IP: address.IP(), Port: int(f.config.ListenPort)}, nil)
	if err != nil {
		return newError("forwarder cannot listen on the port").Base(err)
	}
	go func() {
		err = f.httpserver.Serve(listener)
		if err != nil {
			newError("cannot serve http forward server").Base(err).WriteToLog()
		}
	}()
	return nil
}

func (f *Forwarder) Close() error {
	if f.httpserver != nil {
		return f.httpserver.Close()
	}
	return nil
}

func BridgeResource(rw http.ResponseWriter, r *http.Request, path string) {
	content := path
	if content == "" {
		content = "index.html"
	}
	data, err := securedload.GetAssetSecured("browserforwarder/" + content)
	if err != nil {
		err = newError("cannot load necessary resources").Base(err)
		http.Error(rw, err.Error(), http.StatusForbidden)
		return
	}

	http.ServeContent(rw, r, path, time.Now(), bytes.NewReader(data))
}

func NewForwarder(ctx context.Context, cfg *Config) *Forwarder {
	return &Forwarder{config: cfg, ctx: ctx}
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		return NewForwarder(ctx, cfg.(*Config)), nil
	}))
}
