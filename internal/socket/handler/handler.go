package handler

import "github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"

func InitHandlers() {
	socket.RegisterHandler(socket.Dashboard, Dashboard)
	socket.RegisterHandler(socket.IPDetail, IPDetail)
}
