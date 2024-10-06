package capture

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/ip"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
	"os"
)

// 数据包捕获和抓取

var (
	Handle         *pcap.Handle
	Err            error
	Decoder        gopacket.Decoder
	PacketsCount   int // 总包数
	TrafficCount   int // 总流量
	SessionCount   int // 总会话
	TCPCount       int64
	UDPCount       int64
	OK             bool
	IPEvents       = make(chan ip.PropertyChangeEvent, 100)
	ObserverEvents = make(chan observer.TTLChangeObserverEvent, 100)
)

type Config struct {
	OffLine              string
	Nic                  string
	SnapLen              int32
	BerkeleyPacketFilter string
}

// PacketHandler 处理数据包接口
type PacketHandler interface {
	HandlePacket(packet gopacket.Packet)
}

// StartCapture 开始捕获数据包
func StartCapture(ctx context.Context, c Config, handler PacketHandler, done chan<- struct{}) {
	zap.L().Info(i18n.T("Starting capture"))
	// 启动观察者 goroutine
	zap.L().Info(i18n.T("Starting WatchTTLChange"))
	observer.CleanUp()
	go observer.WatchTTLChange(ObserverEvents)

	// 启动ip属性事件监听 goroutine
	zap.L().Info(i18n.T("Starting ProcessChangeEvent"))
	//for i := 0; i < 10; i++ {
	go ip.ChangeEventIP(IPEvents)
	//}
	zap.L().Debug(i18n.TT("Make Events by Listen IP Change", map[string]interface{}{
		"count": 1000,
	}))
	if c.OffLine != "" {
		Handle, Err = pcap.OpenOffline(c.OffLine)
		zap.L().Info(i18n.TT("Open offline package file", map[string]interface{}{
			"offline": c.OffLine,
		}), zap.Error(Err))
	} else {
		Handle, Err = pcap.OpenLive(c.Nic, c.SnapLen, true, pcap.BlockForever)
		zap.L().Info(i18n.TT("Analyze network card", map[string]interface{}{
			"nic": c.Nic,
		}), zap.Error(Err))
	}

	if c.BerkeleyPacketFilter != "" {
		Err = Handle.SetBPFFilter(c.BerkeleyPacketFilter)
		if Err != nil {
			zap.L().Error("berkeley packet filter panic", zap.Error(Err))
			os.Exit(1)
		}
		zap.L().Info(i18n.TT("Berkeley packet filter set", map[string]interface{}{
			"bpf": c.BerkeleyPacketFilter,
		}))
	}
	if Err != nil {
		zap.L().Error("pcap panic", zap.Error(Err))
		os.Exit(1)
	}

	defer Handle.Close()

	decoderName := fmt.Sprintf("%s", Handle.LinkType())
	if Decoder, OK = gopacket.DecodersByLayerName[decoderName]; !OK {
		panic(fmt.Errorf("decoder %s not found", decoderName))
	}

	source := gopacket.NewPacketSource(Handle, Decoder)
	source.NoCopy = true
	// packet chan
	packets := source.Packets()

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Capture stopped")
			return
		case packet, ok := <-packets:
			if !ok {
				zap.L().Info(i18n.T("Packets Channel Closed"))
				done <- struct{}{}
				return
			}
			PacketsCount++
			handler.HandlePacket(packet)
		}
	}

}
