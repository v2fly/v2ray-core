// +build !confonly

package browserforwarder

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/v2fly/BrowserBridge/handler"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform/securedload"
	"github.com/v2fly/v2ray-core/v4/features/extension"
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
	protocolHeader := false
	protocolHeaderValue := ""
	unsupportedHeader := false
	for k, v := range header {
		if k == "Sec-Websocket-Protocol" {
			protocolHeader = true
			protocolHeaderValue = v[0]
		} else {
			unsupportedHeader = true
		}
	}
	if unsupportedHeader {
		return nil, newError("unsupported header used, only Sec-WebSocket-Protocol is supported for forwarder")
	}
	if !protocolHeader {
		return f.forwarder.Dial(url)
	}
	return f.forwarder.Dial2(url, protocolHeaderValue)
}

func (f *Forwarder) Type() interface{} {
	return extension.BrowserForwarderType()
}

func (f *Forwarder) Start() error {
	if f.config.ListenAddr != "" {
		f.forwarder = handler.NewHttpHandle()
		f.httpserver = &http.Server{Handler: f}

		var listener net.Listener
		var err error
		address := net.ParseAddress(f.config.ListenAddr)

		switch {
		case address.Family().IsIP():
			listener, err = internet.ListenSystem(f.ctx, &net.TCPAddr{IP: address.IP(), Port: int(f.config.ListenPort)}, nil)
		case strings.EqualFold(address.Domain(), "localhost"):
			listener, err = internet.ListenSystem(f.ctx, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: int(f.config.ListenPort)}, nil)
		default:
			return newError("forwarder cannot listen on the address: ", address)
		}
		if err != nil {
			return newError("forwarder cannot listen on the port ", f.config.ListenPort).Base(err)
		}

		go func() {
			if err := f.httpserver.Serve(listener); err != nil {
				newError("cannot serve http forward server").Base(err).WriteToLog()
			}
		}()
	}
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
