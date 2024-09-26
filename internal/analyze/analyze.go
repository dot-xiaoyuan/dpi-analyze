package analyze

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/memory"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"go.uber.org/zap"
	"time"
)

// 流量检测分析

const closeTimeout time.Duration = time.Hour * 24
const timeout time.Duration = time.Minute * 5

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
	streamFactory := &Factory{}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler := reassembly.NewAssembler(streamPool)

	zap.L().Info(i18n.T("Analysis program initialization completed"))

	return &Analyze{
		Assembler: assembler,
	}
}

func (a *Analyze) HandlePacket(packet gopacket.Packet) {
	if packet == nil {
		return
	}
	if packet.NetworkLayer() == nil || packet.TransportLayer() == nil {
		return
	}
	// 累加总流量
	capture.TrafficCount += len(packet.Data())
	// 链路层
	ethernet := capture.Ethernet{}
	if packet.LinkLayer() != nil {
		ethernet.SrcMac = packet.LinkLayer().LinkFlow().Dst().String()
		ethernet.DstMac = packet.LinkLayer().LinkFlow().Src().String()
	}
	// 网络层
	internet := capture.Internet{}
	var ip string
	if packet.NetworkLayer().LayerType() == layers.LayerTypeIPv4 {
		ipv4 := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		// 设置网络层信息 IPv4
		internet.TTL = ipv4.TTL
		internet.DstIP = ipv4.DstIP.String()
		// 设置源IP
		ip = ipv4.SrcIP.String()
	} else if packet.NetworkLayer().LayerType() == layers.LayerTypeIPv6 {
		ipv6 := packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
		// 设置网络层信息 IPv6
		internet.TTL = ipv6.HopLimit
		internet.DstIP = ipv6.DstIP.String()
		// 设置源IP
		ip = ipv6.SrcIP.String()
	}

	// 传输层
	transmission := capture.Transmission{}
	trafficMap := memory.Traffic{Date: time.Now().Format("0102/15/04")}
	if len(packet.TransportLayer().TransportFlow().Src().String()) > len("1024") {
		transmission.UpStream = int64(len(packet.Data()))
	} else {
		transmission.DownStream = int64(len(packet.Data()))
	}
	trafficMap.Update(transmission)

	// 插入 TTL 缓存
	//internetMap := memory.Internet{IP: ip}
	//internetMap.Update(internet)

	// 插入 Mac 缓存
	//ethernetMap := memory.Ethernet{IP: ip}
	//ethernetMap.Update(ethernet)

	// 插入 IP hash 表
	capture.StoreIP(ip, capture.TTL, internet.TTL)
	//capture.StoreIP(ip, capture.Mac, ethernet.SrcMac)

	// analyze TCP
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		ac := &AssemblerContext{
			CaptureInfo: packet.Metadata().CaptureInfo,
		}
		a.Assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, ac)
	}

	if capture.PacketsCount%1000 == 0 {
		ref := packet.Metadata().Timestamp
		_, _ = a.Assembler.FlushWithOptions(reassembly.FlushOptions{
			T:  ref.Add(-timeout),
			TC: ref.Add(-closeTimeout),
		})
	}
}
