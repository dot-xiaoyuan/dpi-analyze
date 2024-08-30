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

type StreamReaderInterface interface {
	GetIdentifier([]byte) ProtocolType
}

type ProtocolHandler interface {
	HandleData(data []byte, reader StreamReaderInterface)
}
