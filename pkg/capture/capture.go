package capture

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
	"sync"
)

// 数据包捕获和抓取

var (
	Decoder      gopacket.Decoder
	PacketsCount int // 总包数
	TrafficCount int // 总流量
	SessionCount int // 总会话
	OK           bool
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
	observer.Setup()

	// 启动ip属性事件监听 goroutine
	zap.L().Info(i18n.T("Starting ProcessChangeEvent"))
	member.Setup()

	var Handle *pcap.Handle
	var Err error
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

		// 获取子网信息
		config.IPNet = utils.GetSubnetInfoByNic(c.Nic)
		zap.L().Info("Listening for packets on interface", zap.String("interface", c.Nic), zap.Any("Network address", config.IPNet))
	}

	if Err != nil {
		zap.L().Error("Failed to open capture device", zap.Error(Err))
		done <- struct{}{}
		return
	}

	if c.BerkeleyPacketFilter != "" {
		Err = Handle.SetBPFFilter(c.BerkeleyPacketFilter)
		if Err != nil {
			zap.L().Error("berkeley packet filter panic", zap.Error(Err))
			done <- struct{}{}
			return
		}
		zap.L().Info(i18n.TT("Berkeley packet filter set", map[string]interface{}{
			"bpf": c.BerkeleyPacketFilter,
		}))
	}
	// 关闭捕获设备
	defer func() {
		if Handle != nil {
			Handle.Close()
			zap.L().Info("Capture device closed")
		}
	}()

	decoderName := fmt.Sprintf("%s", Handle.LinkType())
	if Decoder, OK = gopacket.DecodersByLayerName[decoderName]; !OK {
		zap.L().Fatal("Decoder not found", zap.String("decoder", decoderName))
		return
	}

	source := gopacket.NewPacketSource(Handle, Decoder)
	source.NoCopy = true
	// packet chan
	packets := source.Packets()

	var mu sync.Mutex
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Capture stopped")
			done <- struct{}{}
			return
		case packet, ok := <-packets:
			if !ok {
				zap.L().Info(i18n.T("Packets Channel Closed"))
				done <- struct{}{}
				return
			}
			PacketsCount++
			mu.Lock()
			// 因为需要重组，所以不能使用go协程进行异步处理
			handler.HandlePacket(packet)
			mu.Unlock()
		}
	}
}
