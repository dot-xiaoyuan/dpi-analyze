package protocols

type ProtocolType string

const (
	HTTP    ProtocolType = "http"
	TLS     ProtocolType = "tls"
	UNKNOWN ProtocolType = "unknown"
)
