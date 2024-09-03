package protocols

// ApplicationProtocol 应用层协议
type ApplicationProtocol interface {
	GetProtocol() string
	Parse(data []byte) error
	Handler(data []byte)
}

// ApplicationProtocolData 应用层数据结构
type ApplicationProtocolData struct {
	Protocol string      `bson:"protocols"`
	Data     interface{} `bson:"data"`
}

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
