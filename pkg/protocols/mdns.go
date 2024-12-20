package protocols

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"strings"
)

type Device struct {
	Name        string
	Type        string
	IPv4        string
	IPv6        string
	MAC         string
	Services    string
	Description string // 服务描述（来自 TXT 记录）
}

// 服务类型映射
var serviceDescriptions = map[string]string{
	"_http._tcp.local":           "HTTP 服务",
	"_https._tcp.local":          "HTTPS 服务",
	"_ftp._tcp.local":            "FTP 服务",
	"_ssh._tcp.local":            "SSH 服务",
	"_ipp._tcp.local":            "打印机",
	"_ipps._tcp.local":           "打印机",
	"_airplay._tcp.local":        "AirPlay 服务",
	"_vnc._tcp.local":            "VNC 服务",
	"_printer._tcp.local":        "打印机",
	"_microsoft-ds._tcp.local":   "Windows 文件共享 (SMB)",
	"_sftp-ssh._tcp.local":       "SFTP 服务 (基于 SSH)",
	"_dns-sd._udp.local.":        "用于服务发现的服务类型",
	"_homekit._tcp.local":        "Apple HomeKit 服务",
	"_afpovertcp._tcp.local":     "AFP (Apple Filing Protocol) 文件共享服务",
	"_media._tcp.local":          "媒体流服务",
	"_companion-link._tcp.local": "Apple",
	"_touch-remote._tcp.local":   "Apple",
	"_sync._tcp.local":           "Apple",
	"_daap._tcp.local":           "Apple",
	"_siri._tcp.local":           "Apple",
	"_caldav._tcp.local":         "日历同步服务 (CalDAV)",
	"_carddav._tcp.local":        "联系人同步服务 (CardDAV)",
	"_scp._tcp.local":            "安全复制协议 (SCP)",
	"_hyperion._tcp.local":       "Hyperion 智能灯光系统服务",
	"_xbmc-jsonrpc._tcp.local":   "XBMC (Kodi) JSON-RPC 接口服务",
}

// 根据服务类型返回具体的服务描述
func getServiceDescription(input string) (string, string) {
	// 找到第一个 "._" 出现的位置
	index := strings.Index(input, "._")
	if index == -1 {
		// 如果没有找到 "._" 返回原始字符串和空字符串
		return input, ""
	}

	// 截取分割前后的字符串
	before := input[:index]
	after := input[index+1:] // 跳过 "._" 部分

	return before, serviceDescriptions[after]
}

// ParseMDNS 解析 mDNS 数据包
func ParseMDNS(data []byte, srcIP, srcMac string) Device {
	// 创建 DNS 解析器
	dnsLayer := &layers.DNS{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dnsLayer)
	var decodedLayers []gopacket.LayerType

	// 解码数据包
	err := parser.DecodeLayers(data, &decodedLayers)
	if err != nil {
		fmt.Println("Error decoding mDNS packet:", err)
		return Device{}
	}

	device := Device{
		MAC: srcMac,
	}

	// 使用源 IP 作为默认的设备 IP 地址
	device.IPv4 = srcIP

	// 遍历解码后的层，查找 DNS 层
	for _, layerType := range decodedLayers {
		if layerType == layers.LayerTypeDNS {
			// 解析 Answers 部分（提取设备核心信息）
			if len(dnsLayer.Questions) > 0 {
				return Device{}
			}
			for _, answer := range dnsLayer.Answers {
				if bytes.Contains(answer.Name, []byte("_services._dns")) {
					// 忽略 _services._dns 类型的记录
					continue
				}

				// 只关注有效的记录，忽略 PTR 记录本身，继续处理有效信息
				switch answer.Type {
				case layers.DNSTypePTR:
					device.Name = string(answer.PTR)
				case layers.DNSTypeA:
					// 提取 IPv4 地址
					device.IPv4 = answer.IP.String()
				case layers.DNSTypeAAAA:
					// 提取 IPv6 地址
					device.IPv6 = answer.IP.String()
				case layers.DNSTypeSRV:
					// 提取服务类型信息
					device.Services = string(answer.Name)
				case layers.DNSTypeTXT:
					// 提取设备描述（来自 TXT 记录）
					device.Description = string(answer.TXT)
				}
			}
		}
	}

	if device.Name != "" || device.Services != "" {
		if len(device.Services) > len(device.Name) {
			device.Name = device.Services
		}
		device.Name, device.Type = getServiceDescription(device.Name)
	}
	// 如果设备的 IPv4 或 IPv6 地址有效，则返回该设备
	if len(device.Name) > 0 {
		return device
	}

	// 如果没有有效的 IP 地址信息，返回空设备
	return Device{}
}
