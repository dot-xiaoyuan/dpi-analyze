package types

type Property string

const (
	TTL       Property = "ttl"
	Mac       Property = "mac"
	UserAgent Property = "user-agent"
)

type Feature string

const (
	SNI  Feature = "sni"
	HTTP Feature = "http"
)
