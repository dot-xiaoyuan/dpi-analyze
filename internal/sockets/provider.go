package sockets

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"go.uber.org/zap"
	"net"
	"os"
)

// 数据提供

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type Res provider.Response

func (r *Res) toJson() []byte {
	jsonData, _ := json.Marshal(r)
	return jsonData
}

func NewActionHandler(action string) provider.Handler {
	switch action {
	case "internet":
		return &ActionInternet{}
	case "ethernet":
		return &ActionEthernet{}
	case "dashboard":
		return &ActionDashboard{}
	case "ip":
		return &ActionIP{}
	}
	return nil
}

// capture 通过 unix socket提供数据给前端

func StartUnixSocketServer() {
	// 删除旧的socket TODO sock文件通过配置与flag决定
	_ = os.Remove("/tmp/capture.sock")
	// 创建 socket
	ln, err := net.Listen("unix", "/tmp/capture.sock")
	if err != nil {
		zap.L().Error(i18n.T("Failed create Unix Socket"), zap.Error(err))
		return
	}
	defer ln.Close()

	zap.L().Info(i18n.T("Unix Socket Server listening on /tmp/capture.sock"))

	for {
		conn, err := ln.Accept()
		if err != nil {
			zap.L().Error(i18n.T("Failed accept connection"), zap.Error(err))
			continue
		}

		go handleConnection(conn)
	}
}

// 具体地连接处理
func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			zap.L().Error(i18n.T("Unix socket is closed"), zap.Error(err))
		}
	}(conn)

	res := Res{
		Code:    400,
		Message: "",
	}

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		zap.L().Error("Failed read connection", zap.Error(err))
		res.Message = i18n.T("Failed read connection")
		_, _ = conn.Write(res.toJson())
		return
	}

	params := provider.Request{}
	err = json.Unmarshal(buf[:n], &params)
	if err != nil {
		zap.L().Error(i18n.T("Failed unmarshal params"), zap.Error(err))
		res.Message = i18n.T("Failed unmarshal params")
		_, _ = conn.Write(res.toJson())
		return
	}
	zap.L().Debug(i18n.T("Read connection"), zap.Any("params", params))

	handler := NewActionHandler(params.Action)
	result := handler.Handle(params.Data)
	//if err != nil {
	//	zap.L().Error(i18n.T("Failed read connection"), zap.Error(err))
	//	res.Message = i18n.T("Failed read connection")
	//	conn.Write(res.toJson())
	//	return
	//}
	_, _ = conn.Write(result)
}
