package protocol

// ApplicationProtocol 应用层协议
type ApplicationProtocol interface {
	GetProtocol() string
	Parse(data []byte) error
	Handler(data []byte)
}

// ApplicationProtocolData 应用层数据结构
type ApplicationProtocolData struct {
	Protocol string      `bson:"protocol"`
	Data     interface{} `bson:"data"`
}
