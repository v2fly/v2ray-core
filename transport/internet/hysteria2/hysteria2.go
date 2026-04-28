package hysteria2

import (
	"context"
	"math/rand"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const (
	CanNotUseUDPExtension = "Only hysteria2 proxy protocol can use udpExtension."
)

const (
	closeErrCodeOK            = 0x100 // HTTP3 ErrCodeNoError
	closeErrCodeProtocolError = 0x101 // HTTP3 ErrCodeGeneralProtocolError
	FrameTypeTCPRequest       = 0x401
	udpMessageChanSize        = 1024
	idleCleanupInterval       = 1 * time.Second
	UDPIdleTimeout            = 60 * time.Second
	MaxDatagramFrameSize      = 1200
	URLHost                   = "hysteria"
	URLPath                   = "/auth"
	RequestHeaderAuth         = "Hysteria-Auth"
	ResponseHeaderUDPEnabled  = "Hysteria-UDP"
	CommonHeaderCCRX          = "Hysteria-CC-RX"
	CommonHeaderPadding       = "Hysteria-Padding"
	StatusAuthOK              = 233
)

const (
	paddingChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type padding struct {
	Min int
	Max int
}

func (p padding) String() string {
	n := p.Min + rand.Intn(p.Max-p.Min)
	bs := make([]byte, n)
	for i := range bs {
		bs[i] = paddingChars[rand.Intn(len(paddingChars))]
	}
	return string(bs)
}

var (
	AuthRequestPadding  = padding{Min: 256, Max: 2048}
	AuthResponsePadding = padding{Min: 256, Max: 2048}
	TcpRequestPadding   = padding{Min: 64, Max: 512}
	TcpResponsePadding  = padding{Min: 128, Max: 1024}
)

type datagramKey struct{}

func ContextWithDatagram(ctx context.Context) context.Context {
	return context.WithValue(ctx, datagramKey{}, struct{}{})
}

func DatagramFromContext(ctx context.Context) bool {
	_, ok := ctx.Value(datagramKey{}).(struct{})
	return ok
}

type status int

const (
	StatusNull status = iota
	StatusActive
	StatusInactive
)

const (
	protocolName = "hysteria2"
)

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
