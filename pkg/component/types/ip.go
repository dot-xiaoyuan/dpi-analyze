package types

type Property string

const (
	TTL       Property = "ttl"
	Mac       Property = "mac"
	UserAgent Property = "user_agent"
)

type Feature string

const (
	SNI         Feature = "sni"
	HTTP        Feature = "http"
	TLSVersion  Feature = "tls_version"
	CipherSuite Feature = "cipher_suite"
	Session     Feature = "session"
)
