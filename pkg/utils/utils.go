package utils

import (
	"encoding/binary"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"io"
	"net"
	"strings"
)

// 工具包

func IdentifyClientHello(data []byte) bool {
	if len(data) < 42 {
		return false
	}
	if data[0] != 0x16 {
		return false
	}
	if data[5] != 0x01 {
		return false
	}
	return true
}

// GetServerExtensionName 获取SNI server name indication
func GetServerExtensionName(data []byte) string {
	// Skip past fixed-length records:
	// 1  Handshake Type
	// 3  Length
	// 2  Version (again)
	// 32 Random
	// next Session ID Length
	pos := 38
	dataLen := len(data)

	/* session id */
	if dataLen < pos+1 {
		return ""
	}
	l := int(data[pos])
	pos += l + 1

	/* Cipher Suites */
	if dataLen < (pos + 2) {
		return ""
	}
	l = int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += l + 2

	/* Compression Methods */
	if dataLen < (pos + 1) {
		return ""
	}
	l = int(data[pos])
	pos += l + 1

	/* Extensions */
	if dataLen < (pos + 2) {
		return ""
	}
	extensionsLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2

	/* Parse extensions to get SNI */
	var extensionItemLen int

	/* Parse each 4 bytes for the extension header */
	for pos <= dataLen && pos < extensionsLen+2 {
		if pos+4 > dataLen {
			return ""
		}

		extensionType := binary.BigEndian.Uint16(data[pos : pos+2])
		extensionItemLen = int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		pos += 4

		if pos+extensionItemLen > dataLen {
			return ""
		}

		if extensionType == 0x00 { // SNI extension
			extensionEnd := pos + extensionItemLen
			for pos+3 <= extensionEnd {
				serverNameLen := int(binary.BigEndian.Uint16(data[pos+3 : pos+5]))
				if pos+3+serverNameLen > extensionEnd {
					return ""
				}

				if data[pos] == 0x00 {
					hostname := make([]byte, serverNameLen)
					copy(hostname, data[pos+5:pos+5+serverNameLen])
					return string(hostname)
				}
				// Move to next SNI item
				pos += 3 + serverNameLen
			}
		} else {
			// Move past other extension models
			pos += extensionItemLen
		}
	}
	return ""
}

func GetServerCipherSuite(data []byte) (cipherSuite string) {
	// length = Length(3) + Version(2) + Random(32) + Session ID (1)
	// Skip past fixed-length records:
	// 1  Handshake Type
	// 3  Length
	// 2  Version (again)
	// 32 Random
	// next Session ID Length
	pos := 38
	dataLen := len(data)

	/* session id */
	if dataLen < pos+1 {
		return
	}
	l := int(data[pos])
	pos += l + 1

	/* Cipher Suites */
	if dataLen < (pos + 2) {
		return
	}
	//zap.L().Info("cipherSuite", zap.ByteString("cs", data[pos+2]))
	cs := data[pos : pos+2]
	// zap.L().Info(fmt.Sprintf("0x%02x%02x", cs[0], cs[1]))
	cipherSuite = fmt.Sprintf("0x%02x%02x", cs[0], cs[1])
	return
}

func ReadByConn(conn net.Conn, bufSize int) (data []byte, err error) {
	buffer := make([]byte, 0)        // 用于存放所有数据
	tempBuf := make([]byte, bufSize) // 临时缓冲区

	for {
		n, err := conn.Read(tempBuf)
		if n > 0 {
			buffer = append(buffer, tempBuf[:n]...) // 将读取到的数据追加到总缓冲区中
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed, all data read")
				break // 读取结束
			}
			return nil, fmt.Errorf("error reading: %v", err)
		}
	}
	return buffer, nil
}

func AbsDiff(new, old uint8) uint8 {
	if new > old {
		return new - old
	}
	return old - new
}

func FormatOutput(originText string, width int) string {
	return fmt.Sprintf("%-*s", width, originText)
}

func FormatDomain(sni string) string {
	if strings.Count(sni, ".") >= 2 {
		parts := strings.Split(sni, ".")
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return sni
}

func FormatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetSubnetInfo 获取子网相关信息
func GetSubnetInfo(cidrStr string) (ip net.IP, ipNet *net.IPNet, err error) {
	// 解析 CIDR 地址
	return net.ParseCIDR(cidrStr)
}

// GetBroadcast 获取广播地址
func GetBroadcast(ipNet *net.IPNet) net.IP {
	// 获取广播地址
	broadcast := make(net.IP, len(ipNet.IP))
	for i := range ipNet.IP {
		broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}
	return broadcast
}

func GetIPRange(ipNet *net.IPNet, broadcast net.IP) (net.IP, net.IP) {
	// 有效主机范围
	firstHost := make(net.IP, len(ipNet.IP))
	lastHost := make(net.IP, len(ipNet.IP))
	copy(firstHost, ipNet.IP)
	copy(lastHost, broadcast)
	firstHost[len(firstHost)-1]++ // 第一有效地址
	lastHost[len(lastHost)-1]--   // 最后一有效地址

	return firstHost, lastHost
}

func GetSubnetInfoByNic(nic string) (ip net.IP, ipNet *net.IPNet, err error) {
	// 获取所有网卡接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error fetching interfaces:", err)
	}

	for _, iface := range interfaces {
		if iface.Name != nic {
			continue
		}
		// 获取网卡的所有地址
		var addrs []net.Addr
		addrs, err = iface.Addrs()
		if err != nil {
			return
		}

		// 查找并返回第一个有效地址的子网信息
		for _, addr := range addrs {
			var ok bool
			// 检查地址类型（IPv4 或 IPv6）
			ipNet, ok = addr.(*net.IPNet)
			if !ok {
				continue
			}

			// 跳过链路本地地址（例如 fe80::）
			//if ipNet.IP.IsLinkLocalUnicast() {
			//	continue
			//}
			// 解析并返回 CIDR 信息
			if ipNet, ok = addr.(*net.IPNet); ok {
				return GetSubnetInfo(ipNet.String()) // 调用已有的 GetSubnetInfo 函数
			}
		}
	}
	return
}

// IsIPInRange 检查 IP 地址是否在子网范围内
func IsIPInRange(ip net.IP) bool {
	// 检查 IP 地址是否在子网范围内
	return config.IPNet.Contains(ip)
}
