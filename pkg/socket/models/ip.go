package models

type IPDetail struct {
	QPS      int         `json:"qps"`
	Detail   interface{} `json:"detail"`
	History  `json:"history"`
	Features `json:"features"`
}

type History struct {
	TTL interface{} `json:"ttl"`
	Mac interface{} `json:"mac"`
	Ua  interface{} `json:"ua"`
}

type Features struct {
	SNI         interface{} `json:"sni"`
	HTTP        interface{} `json:"http"`
	TLSVersion  interface{} `json:"tls_version"`
	CipherSuite interface{} `json:"cipher_suite"`
	Session     interface{} `json:"session"`
}
