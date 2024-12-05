package handler

import "github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"

func InitHandlers() {
	socket.RegisterHandler(socket.Dashboard, Dashboard)
	socket.RegisterHandler(socket.IPDetail, IPDetail)
	socket.RegisterHandler(socket.Observer, Observer)
	socket.RegisterHandler(socket.UserList, UserList)
	socket.RegisterHandler(socket.PolicyList, PolicyList)
	socket.RegisterHandler(socket.PolicyUpdate, PolicyUpdate)
	socket.RegisterHandler(socket.ConfigUpdate, ConfigUpdate)
	socket.RegisterHandler(socket.ConfigList, ConfigList)
	socket.RegisterHandler(socket.FeatureLibrary, FeatureLibrary)
	socket.RegisterHandler(socket.FeatureUpdate, FeatureUpdate)
}
