package socket

import (
	"fmt"
)

type MessageType int

const (
	Dashboard = MessageType(iota)
	IPDetail
)

// Message unix 通信数据结构体
type Message struct {
	Type   MessageType `json:"type"`
	Params string      `json:"params"`
}

// MessageHandlerFunc unix 处理方法
type MessageHandlerFunc func(Params string) any

// 全局注册中心，存储消息类型与处理函数的映射
var handlerRegistry = make(map[MessageType]MessageHandlerFunc)

// RegisterHandler 注册消息处理函数
func RegisterHandler(t MessageType, handlerFunc MessageHandlerFunc) {
	handlerRegistry[t] = handlerFunc
}

func (m *Message) handle() (MessageHandlerFunc, error) {
	if handler, exists := handlerRegistry[m.Type]; exists {
		return handler, nil
	}
	return nil, fmt.Errorf("handler not found for message type: %d", m.Type)
}
