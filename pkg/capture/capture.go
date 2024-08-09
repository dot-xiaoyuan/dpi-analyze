package capture

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
	"os"
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
	zap.L().Info("Starting capture")
	if c.OffLine != "" {
		Handle, Err = pcap.OpenOffline(c.OffLine)
		zap.L().Info("pcap open offline", zap.String("OffLine", c.OffLine), zap.Error(Err))
	} else {
		Handle, Err = pcap.OpenLive(c.Nic, c.SnapLen, true, pcap.BlockForever)
		zap.L().Info("pcap open live ", zap.String("Nic", c.Nic), zap.Error(Err))
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
			return
		case packet, ok := <-packets:
			if !ok {
				zap.L().Debug("packets channel closed")
				done <- struct{}{}
				return
			}
			Count++
			handler.HandlePacket(packet)
		}
	}

}
