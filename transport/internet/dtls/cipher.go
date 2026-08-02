package dtls

import piondtls "github.com/pion/dtls/v3"

func configuredCipherSuites(config *Config) []piondtls.CipherSuiteID {
	if config == nil {
		return []piondtls.CipherSuiteID{piondtls.TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256}
	}
	switch config.GetCipherSuite() {
	case DTLSCipherSuite_PSK_AES_128_GCM_SHA256:
		return []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_GCM_SHA256}
	case DTLSCipherSuite_PSK_CHACHA20_POLY1305_SHA256:
		return []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_CHACHA20_POLY1305_SHA256}
	default:
		return []piondtls.CipherSuiteID{piondtls.TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256}
	}
}
