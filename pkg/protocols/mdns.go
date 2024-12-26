package protocols

import (
	"bytes"
	"errors"
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
	parts := strings.SplitN(input, "._", 2)
	if len(parts) < 2 {
		return input, "" // 没有匹配到服务类型，直接返回
	}

	// 查找描述
	description, found := serviceDescriptions["_"+parts[1]]
	if !found {
		return parts[0], "" // 没有匹配到描述，返回基础名称
	}

	return parts[0], description
}

// ParseMDNS 解析 mDNS 数据包
func ParseMDNS(data []byte, srcIP, srcMac string) (Device, error) {
	// 创建 DNS 解析器
	dnsLayer := &layers.DNS{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, dnsLayer)
	var decodedLayers []gopacket.LayerType

	// 解码数据包
	err := parser.DecodeLayers(data, &decodedLayers)
	if err != nil {
		// zap.L().Error("Failed to decode mDNS layer", zap.Error(err))
		return Device{}, err
	}

	device := Device{
		MAC:  srcMac,
		IPv4: srcIP,
	}

	// 遍历解码后的层，查找 DNS 层
	for _, layerType := range decodedLayers {
		if layerType == layers.LayerTypeDNS {
			// 解析 Answers 部分（提取设备核心信息）
			if len(dnsLayer.Questions) > 0 {
				return Device{}, errors.New("DNS questions are not supported")
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

				// 优化：一旦发现关键字段，提前退出循环
				if device.Name != "" && device.IPv4 != "" && device.Services != "" {
					break
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
		return device, nil
	}

	// 如果没有有效的 IP 地址信息，返回空设备
	return Device{}, errors.New("mDNS device not found")
}
