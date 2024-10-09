package socket

import (
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"net"
)

type Request struct {
	Type   MessageType `json:"type"`
	Params interface{} `json:"params"`
}

// SendUnixMessage 向 Unix Socket 服务器发送消息并接收回应
func SendUnixMessage(t MessageType, param interface{}) ([]byte, error) {
	conn, err := net.Dial("unix", config.Cfg.UnixSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %v", err)
	}
	defer conn.Close()

	jsonData, _ := json.Marshal(Request{
		Type:   t,
		Params: param,
	})
	// 发送消息到服务器
	_, err = conn.Write(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	// 读取服务器回应
	data, err := utils.ReadByConn(conn, 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return data, nil
}
