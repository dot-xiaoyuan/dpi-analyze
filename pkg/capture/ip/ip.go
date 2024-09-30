package ip

//
//// IP 相关的核心逻辑
//
//type EventType int
//
//type Property string
//
//const (
//	TTL       Property = "ttl"
//	Mac       Property = "mac"
//	UserAgent Property = "user-agent"
//)
//
//type PropertyChangeEvent struct {
//	IP       string
//	OldValue any
//	NewValue any
//	Property Property
//}
//
//func ChangeEventIP(events <-chan PropertyChangeEvent) {
//	for e := range events {
//		switch e.Property {
//		case TTL:
//			// process ttl
//			break
//		case Mac:
//			// process mac
//			break
//		case UserAgent:
//			// process ua
//			break
//		}
//	}
//}
