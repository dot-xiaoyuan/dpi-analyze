package protocols

// StreamReaderInterface 流Reader接口
type StreamReaderInterface interface {
	GetIdentifier([]byte) ProtocolType
	SetHostName(host string)
	GetIdent() bool
	SetUrls(urls []string)
	GetUrls() []string
	LockParent()
	UnLockParent()
}

type ProtocolHandler interface {
	HandleData(data []byte, reader StreamReaderInterface)
}
