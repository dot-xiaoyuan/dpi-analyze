package analyze

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
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

	logger.Info("Analyze initialized")

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

	// analyze TCP
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		ac := &AssemblerContext{
			CaptureInfo: packet.Metadata().CaptureInfo,
		}
		a.Assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, ac)
	}

	if capture.Count%1000 == 0 {
		ref := packet.Metadata().Timestamp
		_, _ = a.Assembler.FlushWithOptions(reassembly.FlushOptions{
			T:  ref.Add(-timeout),
			TC: ref.Add(-closeTimeout),
		})
	}
}
