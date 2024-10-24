package handler

import "github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"

func InitHandlers() {
	socket.RegisterHandler(socket.Dashboard, Dashboard)
	socket.RegisterHandler(socket.IPDetail, IPDetail)
	socket.RegisterHandler(socket.Observer, Observer)
	socket.RegisterHandler(socket.UserList, UserList)
}
