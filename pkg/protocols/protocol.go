package protocols

// StreamReaderInterface 流Reader接口
type StreamReaderInterface interface {
	GetIdentifier([]byte) ProtocolType
	GetIdent() bool
	SetUrls(urls string)
	GetUrls() []string
	LockParent()
	UnLockParent()
	SetHttpInfo(host, userAgent, contentType, upgrade string)
	SetTlsInfo(sni, version, cipherSuite string)
	SetApplicationProtocol(applicationProtocol ProtocolType)
}

type ProtocolHandler interface {
	HandleData(data []byte, reader StreamReaderInterface) (int, bool)
}
