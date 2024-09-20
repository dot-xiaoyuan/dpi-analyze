package cmd

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/sockets"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/maxmind"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var CaptureCmd = &cobra.Command{
	Use:    "capture",
	Short:  "capture commands",
	PreRun: CaptureRreFunc,
	Run:    CaptureRun,
}

// TODO 流量总览 总流量、总会话数、TCP/UDP会话比列、每秒请求量(RPS)
// TODO IP活动监控 访问量最多的IP、异常TTL改变监控、UA变化趋势、Mac地址变化趋势
// TODO 协议统计 应用层协议分布、TLS版本与加密套件分布
// TODO 流量来源与目的地 源IP和目标IP热图（MaxMind GeoIP）、最频繁访问目标IP
// 初始化Flag
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

// CaptureRreFunc 捕获前置方法
func CaptureRreFunc(c *cobra.Command, args []string) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

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
	// 是否加载geo2ip
	if config.Geo2IP != "" {
		zap.L().Info(i18n.T("Start Load Geo2IP Component"))
		maxmind.Setup(config.Geo2IP)
	}
	// 启动 Unix Socket Server
	go sockets.StartUnixSocketServer()
}

// CaptureRun 捕获子命令入口
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
