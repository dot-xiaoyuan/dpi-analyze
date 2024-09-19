package cmd

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze/cache"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var CaptureCmd = &cobra.Command{
	Use:    "capture",
	Short:  "capture commands",
	PreRun: CaptureRreFunc,
	Run:    CaptureRun,
}

func init() {
	// define flag
	CaptureCmd.Flags().StringVar(&config.CaptureNic, "nic", config.Cfg.Capture.NIC, "capture nic")
	CaptureCmd.Flags().StringVar(&config.CapturePcap, "pcap", config.Cfg.Capture.OfflineFile, "capture pcap file")
	CaptureCmd.Flags().BoolVar(&config.UseMongo, "use-mongo", config.Cfg.UseMongo, "use mongo db")
	CaptureCmd.Flags().BoolVar(&config.ParseFeature, "feature", config.Cfg.ParseFeature, "use parse application")
	CaptureCmd.Flags().StringVar(&config.BerkeleyPacketFilter, "bpf", config.Cfg.BerkeleyPacketFilter, "Berkeley packet filter")
	CaptureCmd.Flags().BoolVar(&config.IgnoreMissing, "ignore-missing", config.Cfg.IgnoreMissing, "ignore missing packet")
	CaptureCmd.Flags().BoolVar(&config.UseTTL, "use-ttl", config.Cfg.UseTTL, "save TTL for IP")
}

func CaptureRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})

	if len(args) == 0 && c.Flags().NFlag() == 0 {
		_ = c.Help()
		os.Exit(0)
	}
	spinners.Setup()
	// 是否使用mongo
	if config.UseMongo {
		zap.L().Info(i18n.T("Start Load Mongodb Component"))
		mongo.Setup()
	}
	// 是否加载特征库
	if config.ParseFeature {
		zap.L().Info(i18n.T("Start Load Feature Component"))
		features.Setup()
	}
	// 启动 Unix Socket Server
	go startUnixSocketServer()
}

func CaptureRun(c *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	// Make Context
	ctx, cancel := context.WithCancel(context.Background())

	// Packet Capture
	assembly := analyze.NewAnalyzer()
	done := make(chan struct{})
	go capture.StartCapture(ctx, capture.Config{
		OffLine:              config.CapturePcap,
		Nic:                  config.CaptureNic,
		SnapLen:              16 << 10,
		BerkeleyPacketFilter: config.BerkeleyPacketFilter,
	}, assembly, done)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-done:
		cancel()
		closed := assembly.Assembler.FlushAll()
		assembly.Factory.WaitGoRoutines()
		zap.L().Info(i18n.TT("Flushed stream", map[string]interface{}{
			"count": closed,
		}))
		//fmt.Println(closed)
	case <-signalChan:
		cancel()
		spinners.Start()
		zap.L().Info(i18n.TT("Received terminate signal, stop analyze...", nil))
		time.Sleep(time.Second)
		spinners.Stop()
		os.Exit(0)
	}
}

func startUnixSocketServer() {
	// 删除旧的socket
	_ = os.Remove("/tmp/capture.sock")
	// 创建 socket
	ln, err := net.Listen("unix", "/tmp/capture.sock")
	if err != nil {
		zap.L().Error("Failed create Unix Socket", zap.Error(err))
		return
	}
	defer ln.Close()

	zap.L().Info("Unix Socket Server listening on /tmp/capture.sock")

	for {
		conn, err := ln.Accept()
		if err != nil {
			zap.L().Error("Failed accept connection", zap.Error(err))
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		zap.L().Error("Failed read connection", zap.Error(err))
		return
	}
	params := strings.TrimSpace(string(buf[:n]))
	zap.L().Debug("Read connection", zap.String("params", params))

	var c capture.LayerMap
	switch params {
	case "internet":
		c = &cache.Internet{}
	case "ethernet":
		c = &cache.Ethernet{}
	}
	all, err := c.QueryAll()
	if err != nil {
		zap.L().Error("Failed read connection", zap.Error(err))
		return
	}
	_, _ = conn.Write(all)
}
