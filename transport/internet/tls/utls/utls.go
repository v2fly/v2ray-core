package utls

import (
	"context"
	systls "crypto/tls"

	utls "github.com/refraction-networking/utls"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func NewUTLSSecurityEngineFromConfig(config *Config) (security.Engine, error) {
	if config.TlsConfig == nil {
		return nil, newError("mandatory field tls_config is not specified")
	}
	return &Engine{config: config}, nil
}

type Engine struct {
	config *Config
}

func (e Engine) Client(conn net.Conn, opts ...security.Option) (security.Conn, error) {
	var options []tls.Option
	for _, v := range opts {
		switch s := v.(type) {
		case security.OptionWithALPN:
			if e.config.ForceAlpn == ForcedALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY {
				options = append(options, tls.WithNextProto(s.ALPNs...))
			}
		case security.OptionWithDestination:
			options = append(options, tls.WithDestination(s.Dest))
		default:
			return nil, newError("unknown option")
		}
	}
	tlsConfig := e.config.TlsConfig.GetTLSConfig(options...)
	utlsConfig, err := uTLSConfigFromTLSConfig(tlsConfig)
	if err != nil {
		return nil, newError("unable to generate utls config from tls config").Base(err)
	}

	preset, err := nameToUTLSPreset(e.config.Imitate)
	if err != nil {
		return nil, newError("unable to get utls preset from name").Base(err)
	}

	utlsClientConn := utls.UClient(conn, utlsConfig, *preset)

	if e.config.NoSNI {
		err = utlsClientConn.RemoveSNIExtension()
		if err != nil {
			return nil, newError("unable to remove server name indication from utls client hello").Base(err)
		}
	}

	err = utlsClientConn.BuildHandshakeState()
	if err != nil {
		return nil, newError("unable to build utls handshake state").Base(err)
	}

	// ALPN is necessary for protocols like websocket to work. The uTLS setting may be overwritten on call into
	// BuildHandshakeState, so we need to check the original tls settings.
	if tlsConfig.NextProtos != nil {
		for n, v := range utlsClientConn.Extensions {
			if aplnExtension, ok := v.(*utls.ALPNExtension); ok {
				if e.config.ForceAlpn == ForcedALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY {
					aplnExtension.AlpnProtocols = tlsConfig.NextProtos
					break
				}
				if e.config.ForceAlpn == ForcedALPN_NO_ALPN {
					utlsClientConn.Extensions = append(utlsClientConn.Extensions[:n], utlsClientConn.Extensions[n+1:]...)
					break
				}
			}
		}
	}

	err = utlsClientConn.BuildHandshakeState()
	if err != nil {
		return nil, newError("unable to build utls handshake state after modification").Base(err)
	}

	err = utlsClientConn.Handshake()
	if err != nil {
		return nil, newError("unable to finish utls handshake").Base(err)
	}
	return uTLSClientConnection{utlsClientConn}, nil
}

type uTLSClientConnection struct {
	*utls.UConn
}

func (u uTLSClientConnection) GetConnectionApplicationProtocol() (string, error) {
	if err := u.Handshake(); err != nil {
		return "", err
	}
	return u.ConnectionState().NegotiatedProtocol, nil
}

func uTLSConfigFromTLSConfig(config *systls.Config) (*utls.Config, error) { // nolint: unparam
	uconfig := &utls.Config{
		Rand:                  config.Rand,
		Time:                  config.Time,
		RootCAs:               config.RootCAs,
		NextProtos:            config.NextProtos,
		ServerName:            config.ServerName,
		VerifyPeerCertificate: config.VerifyPeerCertificate,
		InsecureSkipVerify:    config.InsecureSkipVerify,
		ClientAuth:            utls.ClientAuthType(config.ClientAuth),
		ClientCAs:             config.ClientCAs,
	}
	return uconfig, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewUTLSSecurityEngineFromConfig(config.(*Config))
	}))
}
