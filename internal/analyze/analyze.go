package analyze

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/resolve"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/protocols"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/sessions"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/statictics"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"go.uber.org/zap"
	"net"
	"strings"
	"time"
)

// 流量检测分析

const closeTimeout time.Duration = time.Second * 30
const timeout time.Duration = time.Second * 10

type Analyze struct {
	Assembler *reassembly.Assembler
	Factory   Factory
}

// AssemblerContext provides method to get metadata
type AssemblerContext struct {
	CaptureInfo gopacket.CaptureInfo
}

func (ac *AssemblerContext) GetCaptureInfo() gopacket.CaptureInfo {
	return ac.CaptureInfo
}

func NewAnalyzer() *Analyze {
	//
	sessions.StartLogConsumer()
	resolve.StartUserAgentConsumer()
	// 清空有序集合以及遗留数据
	member.CleanUp()
	streamFactory := &Factory{}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)

	zap.L().Info(i18n.T("Analysis program initialization completed"))

	return &Analyze{
		Assembler: assembler,
	}
}

func (a *Analyze) HandlePacket(packet gopacket.Packet) {
	if packet.NetworkLayer() == nil || packet.TransportLayer() == nil {
		return
	}
	// 累加总流量
	capture.TrafficCount += len(packet.Data())
	// 链路层
	ethernet := types.Ethernet{}
	if packet.LinkLayer() != nil {
		ethernet.SrcMac = packet.LinkLayer().LinkFlow().Src().String()
		ethernet.DstMac = packet.LinkLayer().LinkFlow().Dst().String()
	}
	// 网络层
	internet := types.Internet{}
	var ip, dip string
	var srcIPNet net.IP
	if packet.NetworkLayer().LayerType() == layers.LayerTypeIPv4 {
		ipv4 := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		// 设置网络层信息 IPv4
		internet.TTL = ipv4.TTL
		internet.DstIP = ipv4.DstIP.String()
		// 设置源IP
		srcIPNet = ipv4.SrcIP
		ip = ipv4.SrcIP.String()
		dip = ipv4.DstIP.String()
	} else if packet.NetworkLayer().LayerType() == layers.LayerTypeIPv6 {
		ipv6 := packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
		// 设置网络层信息 IPv6
		internet.TTL = ipv6.HopLimit
		internet.DstIP = ipv6.DstIP.String()
		// 设置源IP
		srcIPNet = ipv6.SrcIP
		ip = ipv6.SrcIP.String()
		dip = ipv6.DstIP.String()
	}

	// 过滤广播地址
	//if ip == "255.255.255.255" {
	//	return
	//}
	var srcPort, dstPort string
	if packet.TransportLayer() != nil {
		srcPort, dstPort = packet.TransportLayer().TransportFlow().Src().String(), packet.TransportLayer().TransportFlow().Dst().String()
	} else {
		srcPort, dstPort = "", ""
	}

	// user_ip 转储缓存
	var userIP, tranIP, userMac string
	if users.ExitsUser(ip) {
		userIP, tranIP, userMac = ip, dip, ethernet.SrcMac
	} else if users.ExitsUser(dip) {
		userIP, tranIP, userMac = dip, ip, ethernet.DstMac
	}
	// 仅关注在线用户 如果在线用户中不存在该IP跳过该数据包
	if config.FollowOnlyOnlineUsers {
		if userIP == "" {
			return
		}
	} else {
		if len(config.IPNet) > 0 && utils.IsIPInRange(srcIPNet) {
			userIP, tranIP, userMac = ip, dip, ethernet.SrcMac
		} else {
			return
		}
	}
	// mDNS
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		if (srcPort == "5353" || dstPort == "5353") &&
			packet.NetworkLayer().LayerType() == layers.LayerTypeIPv4 &&
			packet.ApplicationLayer() != nil {

			payload := packet.ApplicationLayer().Payload()
			device, err := protocols.ParseMDNS(payload, userIP, userMac)
			if err == nil {
				if len(device.Name) > 0 || len(device.Type) > 0 || len(device.MAC) > 0 {
					_ = ants.Submit(func() {
						//zap.L().Debug("mDNS", zap.Any("device", device))
						if len(strings.TrimSpace(device.Name)) > 0 {
							member.Store(member.Hash{
								IP:    userIP,
								Field: types.DeviceName,
								Value: device.Name,
							})
						}
						if len(strings.TrimSpace(device.Type)) > 0 {
							member.Store(member.Hash{
								IP:    userIP,
								Field: types.DeviceType,
								Value: device.Type,
							})
						}
						if len(device.MAC) > 0 { // 假设 MAC 校验在 ParseMDNS 已完成
							member.Store(member.Hash{
								IP:    userIP,
								Field: types.Mac,
								Value: device.MAC,
							})
						}
					})
				}
			}
		}
	}

	// 记录mac和ip地址绑定关系
	// 如果 TTL = 255，跳过该数据包
	if internet.TTL == 255 {
		return
	}

	if transportType := packet.TransportLayer().LayerType().String(); transportType != "" {
		if transportType == "TCP" {
			statictics.TransportLayer.Increment("tcp")
		} else if transportType == "UDP" {
			statictics.TransportLayer.Increment("udp")
		}
	}
	// 传输层
	transmission := types.Transmission{}
	trafficMap := memory.Traffic{Date: time.Now().Format("01-02/15/04")}
	if len(packet.TransportLayer().TransportFlow().Src().String()) > len("1024") {
		transmission.UpStream = int64(len(packet.Data()))
	} else {
		transmission.DownStream = int64(len(packet.Data()))
	}
	trafficMap.Update(transmission)

	if config.UseTTL && userIP == ip {
		_ = ants.Submit(func() { // 插入 IP hash TTL表
			member.Store(member.Hash{
				IP:    userIP,
				Field: types.TTL,
				Value: internet.TTL,
			})
			// 如果TTL = 127，则记录设备为win
			if internet.TTL >= 126 && internet.TTL <= 128 {
				// 判断缓存中是否有
				if !member.GetAnalyze(ip) {
					resolve.AnalyzeByTTL(userIP, internet.TTL)
					member.PutAnalyze(ip)
				}
			}
		})
	}

	if len(userMac) > 0 && userIP == ip {
		_ = ants.Submit(func() { // 插入 IP hash Mac表
			member.Store(member.Hash{
				IP:    userIP,
				Field: types.Mac,
				Value: userMac,
			})
		})
	}

	// analyze TCP
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		ac := &AssemblerContext{
			CaptureInfo: packet.Metadata().CaptureInfo,
		}
		a.Assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, ac)
	}
	// analyze UDP
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)

		layerType := CheckUDP(userIP, tranIP, udp)
		// dhcp协议日志输出
		//if layerType == layers.LayerTypeDHCPv4 {
		//	dhcp := packet.Layer(layers.LayerTypeDHCPv4).(*layers.DHCPv4)
		//	zap.L().Debug("dhcp", zap.Any("layer", dhcp))
		//}
		if layerType == layers.LayerTypeDNS {
			dnsLayer := packet.Layer(layers.LayerTypeDNS)
			if dnsLayer != nil {
				dns := dnsLayer.(*layers.DNS)
				for _, quest := range dns.Questions {
					if ok, domain := features.HandleFeatureMatch(string(quest.Name), userIP, types.DeviceRecord{}); ok {
						zap.L().Debug("DNS Question", zap.String("quest", fmt.Sprintf("%s", quest.Name)), zap.String("ip", userIP))
						//zap.L().Debug("DNS Question", zap.String("Question", fmt.Sprintf("%s", quest.Name)), zap.String("domain", domain.BrandName))
						resolve.Handle(types.DeviceRecord{
							IP:           userIP,
							OriginChanel: types.DNSProperty,
							OriginValue:  string(quest.Name),
							Os:           "",
							Version:      "",
							Device:       "",
							Brand:        strings.ToLower(domain.BrandName),
							Model:        "",
							Description:  domain.Description,
							Icon:         domain.Icon,
							LastSeen:     time.Now(),
						})
					}
				}
			}
		}
		// 会话数累加
		capture.SessionCount++

		_ = ants.Submit(func() {
			member.Increment(types.Feature{ // 会话数
				IP:    userIP,
				Field: types.Session,
				Value: tranIP,
			})
		})
	}

	if capture.PacketsCount%100000 == 0 {
		//zap.L().Debug(i18n.T("capture packet"), zap.Int("count", capture.PacketsCount))
		ref := packet.Metadata().Timestamp
		//zap.L().Debug("metadata timestamp", zap.Any("ref", ref))
		flushed, closed := a.Assembler.FlushWithOptions(reassembly.FlushOptions{
			T:  ref.Add(-time.Minute),
			TC: ref.Add(-time.Minute * 2),
		})
		zap.L().Debug("Flush stream", zap.Int("flushed", flushed), zap.Int("closed", closed))
	}
}

//func (a *Analyze) FlushWithOptions() {
//	a.Assembler.FlushCloseOlderThan(time.Now().Add(-time.Minute * 5))
//}

func (a *Analyze) FlushStream(ctx context.Context) {
	// 定期刷新流的 Goroutine
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Stopping flow assembler flush")
			return
		case <-ticker.C:
			a.Assembler.FlushCloseOlderThan(time.Now().Add(-time.Minute))
		}
	}
}
