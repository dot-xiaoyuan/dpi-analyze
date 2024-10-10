package socket

import (
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"go.uber.org/zap"
	"net"
	"os"
)

// StartServer 启动 Unix Socket 服务器
func StartServer() {
	_ = os.Remove(config.Cfg.UnixSocket) // 清理旧的 socket 文件
	l, err := net.Listen("unix", config.Cfg.UnixSocket)
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to listen on socket: %v", err))
		os.Exit(1)
	}
	zap.L().Info(i18n.TT("Unix Socket Server listening", map[string]interface{}{
		"sock": config.Cfg.UnixSocket,
	}))
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("read error:", err)
		return
	}

	// 处理客户端发送的消息
	zap.L().Debug("Received message", zap.String("message", string(buf[:n])))

	var req Message
	err = json.Unmarshal(buf[:n], &req)
	if err != nil {
		fmt.Println("unmarshal error:", err)
		return
	}
	// 获取处理函数
	handler, err := req.handle()
	if err != nil {
		_, _ = conn.Write([]byte("Unknown message type"))
	}

	response := handler(req.Params)
	var res []byte
	res, _ = json.Marshal(response)
	_, _ = conn.Write(res)
}
