package protocols

import (
	"crypto/md5"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/statictics"
	"regexp"
)

var (
	httpRequestPattern  = regexp.MustCompile(`^(GET|POST|PUT|DELETE|HEAD|OPTIONS)`)
	httpResponsePattern = regexp.MustCompile(`^HTTP/1.`)
)

// IdentifyProtocol 识别协议
func IdentifyProtocol(buffer []byte, srcPort, dstPort string) ProtocolType {
	// 统计应用层数
	if srcPort == "80" || dstPort == "80" || checkHttp(buffer) {
		// 统计http数
		statictics.ApplicationLayer.Increment("http")
		return HTTP
	}
	if srcPort == "443" || dstPort == "443" {
		// 统计
		statictics.ApplicationLayer.Increment("tls")
		return TLS
	}
	if srcPort == "53" || dstPort == "53" {
		// 统计 DNS
		statictics.ApplicationLayer.Increment("dns")
		return DNS
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
	if CheckHttpByRequest(buffer) || CheckHttpByResponse(buffer) {
		return true
	}
	return false
}

// CheckHttpByRequest check 是否是http请求
func CheckHttpByRequest(data []byte) bool {
	return httpRequestPattern.Match(data)
}

// CheckHttpByResponse check 是否是http响应
func CheckHttpByResponse(data []byte) bool {
	return httpResponsePattern.Match(data)
}
