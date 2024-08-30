package protocols

func IdentifyProtocol(buffer []byte, srcPort, dstPort string) ProtocolType {
	if srcPort == "80" || dstPort == "80" {
		return HTTP
	}
	if srcPort == "443" || dstPort == "443" {
		return TLS
	}
	return UNKNOWN
}
