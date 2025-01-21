package tls

import (
	"crypto/hmac"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol/tls/cert"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

var globalSessionCache = tls.NewLRUClientSessionCache(128)

const exp8357 = "experiment:8357"

// ParseCertificate converts a cert.Certificate to Certificate.
func ParseCertificate(c *cert.Certificate) *Certificate {
	if c != nil {
		certPEM, keyPEM := c.ToPEM()
		return &Certificate{
			Certificate: certPEM,
			Key:         keyPEM,
		}
	}
	return nil
}

func (c *Config) loadSelfCertPool(usage Certificate_Usage) (*x509.CertPool, error) {
	root := x509.NewCertPool()
	for _, cert := range c.Certificate {
		if cert.Usage == usage {
			if !root.AppendCertsFromPEM(cert.Certificate) {
				return nil, newError("failed to append cert").AtWarning()
			}
		}
	}
	return root, nil
}

// BuildCertificates builds a list of TLS certificates from proto definition.
func (c *Config) BuildCertificates() []tls.Certificate {
	certs := make([]tls.Certificate, 0, len(c.Certificate))
	for _, entry := range c.Certificate {
		if entry.Usage != Certificate_ENCIPHERMENT {
			continue
		}
		keyPair, err := tls.X509KeyPair(entry.Certificate, entry.Key)
		if err != nil {
			newError("ignoring invalid X509 key pair").Base(err).AtWarning().WriteToLog()
			continue
		}
		certs = append(certs, keyPair)
	}
	return certs
}

func isCertificateExpired(c *tls.Certificate) bool {
	if c.Leaf == nil && len(c.Certificate) > 0 {
		if pc, err := x509.ParseCertificate(c.Certificate[0]); err == nil {
			c.Leaf = pc
		}
	}

	// If leaf is not there, the certificate is probably not used yet. We trust user to provide a valid certificate.
	return c.Leaf != nil && c.Leaf.NotAfter.Before(time.Now().Add(time.Minute*2))
}

func issueCertificate(rawCA *Certificate, domain string) (*tls.Certificate, error) {
	parent, err := cert.ParseCertificate(rawCA.Certificate, rawCA.Key)
	if err != nil {
		return nil, newError("failed to parse raw certificate").Base(err)
	}
	newCert, err := cert.Generate(parent, cert.CommonName(domain), cert.DNSNames(domain))
	if err != nil {
		return nil, newError("failed to generate new certificate for ", domain).Base(err)
	}
	newCertPEM, newKeyPEM := newCert.ToPEM()
	cert, err := tls.X509KeyPair(newCertPEM, newKeyPEM)
	return &cert, err
}

func (c *Config) getCustomCA() []*Certificate {
	certs := make([]*Certificate, 0, len(c.Certificate))
	for _, certificate := range c.Certificate {
		if certificate.Usage == Certificate_AUTHORITY_ISSUE {
			certs = append(certs, certificate)
		}
	}
	return certs
}

func getGetCertificateFunc(c *tls.Config, ca []*Certificate) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var access sync.RWMutex

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		certExpired := false

		access.RLock()
		certificate, found := c.NameToCertificate[domain]
		access.RUnlock()

		if found {
			if !isCertificateExpired(certificate) {
				return certificate, nil
			}
			certExpired = true
		}

		if certExpired {
			newCerts := make([]tls.Certificate, 0, len(c.Certificates))

			access.Lock()
			for _, certificate := range c.Certificates {
				cert := certificate
				if !isCertificateExpired(&cert) {
					newCerts = append(newCerts, cert)
				} else if cert.Leaf != nil {
					expTime := cert.Leaf.NotAfter.Format(time.RFC3339)
					newError("old certificate for ", domain, " (expire on ", expTime, ") discard").AtInfo().WriteToLog()
				}
			}

			c.Certificates = newCerts
			access.Unlock()
		}

		var issuedCertificate *tls.Certificate

		// Create a new certificate from existing CA if possible
		for _, rawCert := range ca {
			if rawCert.Usage == Certificate_AUTHORITY_ISSUE {
				newCert, err := issueCertificate(rawCert, domain)
				if err != nil {
					newError("failed to issue new certificate for ", domain).Base(err).WriteToLog()
					continue
				}
				parsed, err := x509.ParseCertificate(newCert.Certificate[0])
				if err == nil {
					newCert.Leaf = parsed
					expTime := parsed.NotAfter.Format(time.RFC3339)
					newError("new certificate for ", domain, " (expire on ", expTime, ") issued").AtInfo().WriteToLog()
				} else {
					newError("failed to parse new certificate for ", domain).Base(err).WriteToLog()
				}

				access.Lock()
				c.Certificates = append(c.Certificates, *newCert)
				issuedCertificate = &c.Certificates[len(c.Certificates)-1]
				access.Unlock()
				break
			}
		}

		if issuedCertificate == nil {
			return nil, newError("failed to create a new certificate for ", domain)
		}

		access.Lock()
		c.BuildNameToCertificate()
		access.Unlock()

		return issuedCertificate, nil
	}
}

func (c *Config) IsExperiment8357() bool {
	return strings.HasPrefix(c.ServerName, exp8357)
}

func (c *Config) parseServerName() string {
	if c.IsExperiment8357() {
		return c.ServerName[len(exp8357):]
	}

	return c.ServerName
}

func (c *Config) verifyPeerCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if c.PinnedPeerCertificateChainSha256 != nil {
		hashValue := GenerateCertChainHash(rawCerts)
		for _, v := range c.PinnedPeerCertificateChainSha256 {
			if hmac.Equal(hashValue, v) {
				return nil
			}
		}
		return newError("peer cert is unrecognized: ", base64.StdEncoding.EncodeToString(hashValue))
	}
	return nil
}

type alwaysFlushWriter struct {
	file *os.File
}

func (a *alwaysFlushWriter) Write(p []byte) (n int, err error) {
	n, err = a.file.Write(p)
	a.file.Sync()
	return n, err
}

// GetTLSConfig converts this Config into tls.Config.
func (c *Config) GetTLSConfig(opts ...Option) *tls.Config {
	root, err := c.getCertPool()
	if err != nil {
		newError("failed to load system root certificate").AtError().Base(err).WriteToLog()
	}

	if c == nil {
		return &tls.Config{
			ClientSessionCache:     globalSessionCache,
			RootCAs:                root,
			InsecureSkipVerify:     false,
			NextProtos:             nil,
			SessionTicketsDisabled: true,
		}
	}

	clientRoot, err := c.loadSelfCertPool(Certificate_AUTHORITY_VERIFY_CLIENT)
	if err != nil {
		newError("failed to load client root certificate").AtError().Base(err).WriteToLog()
	}

	config := &tls.Config{
		ClientSessionCache:     globalSessionCache,
		RootCAs:                root,
		InsecureSkipVerify:     c.AllowInsecure,
		NextProtos:             c.NextProtocol,
		SessionTicketsDisabled: !c.EnableSessionResumption,
		VerifyPeerCertificate:  c.verifyPeerCert,
		ClientCAs:              clientRoot,
	}

	if c.AllowInsecureIfPinnedPeerCertificate && c.PinnedPeerCertificateChainSha256 != nil {
		config.InsecureSkipVerify = true
	}

	for _, opt := range opts {
		opt(config)
	}

	config.Certificates = c.BuildCertificates()
	config.BuildNameToCertificate()

	caCerts := c.getCustomCA()
	if len(caCerts) > 0 {
		config.GetCertificate = getGetCertificateFunc(config, caCerts)
	}

	if sn := c.parseServerName(); len(sn) > 0 {
		config.ServerName = sn
	}

	if len(config.NextProtos) == 0 {
		config.NextProtos = []string{"h2", "http/1.1"}
	}

	if c.VerifyClientCertificate {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	switch c.MinVersion {
	case Config_TLS1_0:
		config.MinVersion = tls.VersionTLS10
	case Config_TLS1_1:
		config.MinVersion = tls.VersionTLS11
	case Config_TLS1_2:
		config.MinVersion = tls.VersionTLS12
	case Config_TLS1_3:
		config.MinVersion = tls.VersionTLS13
	}

	switch c.MaxVersion {
	case Config_TLS1_0:
		config.MaxVersion = tls.VersionTLS10
	case Config_TLS1_1:
		config.MaxVersion = tls.VersionTLS11
	case Config_TLS1_2:
		config.MaxVersion = tls.VersionTLS12
	case Config_TLS1_3:
		config.MaxVersion = tls.VersionTLS13
	}

	if len(c.EchConfig) > 0 || len(c.Ech_DOHserver) > 0 {
		err := ApplyECH(c, config) //nolint: staticcheck
		if err != nil {            //nolint: staticcheck
			newError("unable to set ECH").AtError().Base(err).WriteToLog()
		}
	}

	return config
}

// Option for building TLS config.
type Option func(*tls.Config)

// WithDestination sets the server name in TLS config.
func WithDestination(dest net.Destination) Option {
	return func(config *tls.Config) {
		if config.ServerName == "" {
			switch dest.Address.Family() {
			case net.AddressFamilyDomain:
				config.ServerName = dest.Address.Domain()
			case net.AddressFamilyIPv4, net.AddressFamilyIPv6:
				config.ServerName = dest.Address.IP().String()
			}
		}
	}
}

// WithNextProto sets the ALPN values in TLS config.
func WithNextProto(protocol ...string) Option {
	return func(config *tls.Config) {
		if len(config.NextProtos) == 0 {
			config.NextProtos = protocol
		}
	}
}

// ConfigFromStreamSettings fetches Config from stream settings. Nil if not found.
func ConfigFromStreamSettings(settings *internet.MemoryStreamConfig) *Config {
	if settings == nil {
		return nil
	}
	if settings.SecuritySettings == nil {
		return nil
	}
	// Fail close for unknown TLS settings type.
	// For TLS Clients, Security Engine should be used, instead of this.
	config := settings.SecuritySettings.(*Config)
	return config
}
