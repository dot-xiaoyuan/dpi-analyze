package protocols

import (
	"crypto/md5"
	"fmt"
)

// IdentifyProtocol 识别协议
func IdentifyProtocol(buffer []byte, srcPort, dstPort string) ProtocolType {
	if srcPort == "80" || dstPort == "80" {
		return HTTP
	}
	if srcPort == "443" || dstPort == "443" {
		return TLS
	}
	return UNKNOWN
}

// GenerateSessionId 生成五元祖hash
func GenerateSessionId(srcIP, dstIP, srcPort, dstPort, protocol string) string {
	plant := fmt.Sprintf("%s%s%s%s%s", srcIP, dstIP, srcPort, dstPort, protocol)
	return fmt.Sprintf("%x", md5.Sum([]byte(plant)))
}
