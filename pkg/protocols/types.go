package protocols

type ProtocolType string

const (
	HTTP    ProtocolType = "http"
	TLS     ProtocolType = "tls"
	DNS     ProtocolType = "dns"
	UNKNOWN ProtocolType = "unknown"
)
