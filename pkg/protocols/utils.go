package protocols

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

// IdentifyProtocol 识别协议
func IdentifyProtocol(buffer []byte, srcPort, dstPort string) ProtocolType {
	if srcPort == "80" || dstPort == "80" || checkHttp(buffer) {
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

func checkHttp(buffer []byte) bool {
	//fmt.Printf("%s", hex.Dump(buffer))
	payload := string(buffer)

	if CheckHttpByRequest(payload) || CheckHttpByResponse(payload) {
		return true
	}
	return false
}

// CheckHttpByRequest check 是否是http请求
func CheckHttpByRequest(payload string) bool {
	if strings.HasPrefix(payload, http.MethodGet) ||
		strings.HasPrefix(payload, http.MethodPost) ||
		strings.HasPrefix(payload, http.MethodPut) ||
		strings.HasPrefix(payload, http.MethodDelete) ||
		strings.HasPrefix(payload, http.MethodOptions) ||
		strings.HasPrefix(payload, http.MethodHead) {
		return true
	}
	return false
}

// CheckHttpByResponse check 是否是http响应
func CheckHttpByResponse(payload string) bool {
	if strings.HasPrefix(payload, "HTTP/1.") {
		return true
	}
	return false
}
