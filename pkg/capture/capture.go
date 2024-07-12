package capture

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
)

// 数据包捕获和抓取

var (
	Handle  *pcap.Handle
	Err     error
	Decoder gopacket.Decoder
	Count   int
	OK      bool
)

type Config struct {
	OffLine string
	Nic     string
	SnapLen int32
}

// PacketHandler 处理数据包接口
type PacketHandler interface {
	HandlePacket(packet gopacket.Packet)
}

// StartCapture 开始捕获数据包
func StartCapture(ctx context.Context, c Config, handler PacketHandler, done chan<- struct{}) {
	logger.Info("Starting capture")
	if c.OffLine != "" {
		Handle, Err = pcap.OpenOffline(c.OffLine)
		logger.Info("pcap open offline", zap.String("OffLine", c.OffLine), zap.Error(Err))
	} else {
		Handle, Err = pcap.OpenLive(c.Nic, c.SnapLen, true, pcap.BlockForever)
		logger.Info("pcap open live ", zap.String("Nic", c.Nic), zap.Error(Err))
	}

	defer Handle.Close()

	if Err != nil {
		panic(Err)
	}

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
			return
		case packet, ok := <-packets:
			if !ok {
				logger.Debug("packets channel closed")
				done <- struct{}{}
				return
			}
			Count++
			handler.HandlePacket(packet)
		}
	}

}
