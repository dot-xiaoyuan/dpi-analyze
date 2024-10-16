package cmd

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/socket/handler"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/traffic"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/cron"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/maxmind"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/uaparser"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/sevlyar/go-daemon"
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

var captureDaemon = &utils.Daemon{
	Name: "capture",
	Context: daemon.Context{
		PidFileName: fmt.Sprintf("%s/capture.pid", config.RunDir),
		PidFilePerm: 0644,
		LogFileName: fmt.Sprintf("%s/capture.log", config.LogDir),
		LogFilePerm: 0640,
		WorkDir:     config.Home,
		Args:        os.Args,
		Umask:       027,
	},
}

// TODO 流量总览 总流量、总会话数、TCP/UDP会话比列、每秒请求量(RPS)
// TODO IP活动监控 访问量最多的IP、异常TTL改变监控、UA变化趋势、Mac地址变化趋势
// TODO 协议统计 应用层协议分布、TLS版本与加密套件分布
// TODO 流量来源与目的地 源IP和目标IP热图（MaxMind GeoIP）、最频繁访问目标IP
// 初始化Flag
func init() {
	// 初始化加载组件
	spinners.Setup()
	// define flag
	CaptureCmd.Flags().StringVar(&config.CaptureNic, "nic", config.Cfg.Capture.NIC, "capture nic")
	CaptureCmd.Flags().StringVar(&config.CapturePcap, "pcap", config.Cfg.Capture.OfflineFile, "capture pcap file")
	CaptureCmd.Flags().BoolVar(&config.UseMongo, "use-mongo", config.Cfg.UseMongo, "use mongo db")
	CaptureCmd.Flags().BoolVar(&config.ParseFeature, "feature", config.Cfg.ParseFeature, "use parse application")
	CaptureCmd.Flags().StringVar(&config.BerkeleyPacketFilter, "bpf", config.Cfg.BerkeleyPacketFilter, "Berkeley packet filter")
	CaptureCmd.Flags().BoolVar(&config.IgnoreMissing, "ignore-missing", config.Cfg.IgnoreMissing, "ignore missing packet")
	CaptureCmd.Flags().BoolVar(&config.FollowOnlyOnlineUsers, "follow-online-users", config.Cfg.FollowOnlyOnlineUsers, "follow only online users")
	CaptureCmd.Flags().BoolVar(&config.UseTTL, "use-ttl", config.Cfg.UseTTL, "save TTL for IP")
	CaptureCmd.Flags().BoolVar(&config.UseUA, "use-ua", config.Cfg.UseUA, "use ua parser")
	CaptureCmd.Flags().StringVar(&config.Geo2IP, "geo2ip", config.Cfg.Geo2IP, "geo2ip")
}

// CaptureRreFunc 捕获前置方法
func CaptureRreFunc(c *cobra.Command, args []string) {

	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})

	if len(args) == 0 && c.Flags().NFlag() == 0 {
		_ = c.Help()
		os.Exit(0)
	}
}

// CaptureRun 捕获子命令入口
func CaptureRun(*cobra.Command, []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	if config.Signal != "" {
		switch config.Signal {
		case types.STOP:
			captureDaemon.Stop()
		case types.START:
			captureDaemon.Start(captureRun)
		case types.STATUS:
			captureDaemon.Status()
		case types.RESTART:
			captureDaemon.Restart(captureRun)
		default:
			fmt.Println("Usage: [start|stop|status|restart]")
		}
		os.Exit(0)
	}
	if config.Detach {
		captureDaemon.Start(captureRun)
		return
	}
	captureRun()
}

func captureRun() {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动系统信号监听
	go handleSignals(cancel)

	// 使用spinner加载组件
	loadComponents()

	// 启动 Packet Capture
	assembly := analyze.NewAnalyzer()
	done := make(chan struct{})

	go capture.StartCapture(ctx, capture.Config{
		OffLine:              config.CapturePcap,
		Nic:                  config.CaptureNic,
		SnapLen:              16 << 10,
		BerkeleyPacketFilter: config.BerkeleyPacketFilter,
	}, assembly, done)

	// wait signal
	select {
	case <-done:
		handleCaptureCompletion(cancel, assembly)
	}
}

// 信号监听 Goroutine
func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-signalChan
	zap.L().Info("Received signal", zap.String("signal", sig.String()))
	handleGracefulExit(cancel)
}

// 加载所有组件并使用 Spinner 提示
func loadComponents() {
	spinners.WithSpinner("Loading Redis Component", redis.Setup)
	spinners.WithSpinner("Loading Cron  Component", cron.Setup)
	spinners.WithSpinner("Loading Ants  Component", func() error {
		return ants.Setup(100)
	})

	if config.UseMongo {
		spinners.WithSpinner("Loading MongoDB Component", mongo.Setup)
	}

	if config.ParseFeature {
		spinners.WithSpinner("Loading Feature Component", features.Setup)
	}

	if config.UseUA {
		spinners.WithSpinner("Loading UserAgent Component", uaparser.Setup)
	}

	if config.Geo2IP != "" {
		spinners.WithSpinner("Loading Geo2IP Component", func() error {
			return maxmind.Setup(config.Geo2IP)
		})
	}

	if err := ants.Submit(socket.StartServer); err != nil {
		zap.L().Error("Failed to start unix sock server", zap.Error(err))
		os.Exit(1)
	}

	handler.InitHandlers()
	userSync := users.UserSync{}
	userSync.CleanUp()

	_, err := cron.AddJob("@every 1m", userSync)
	if err != nil {
		zap.L().Error("Failed to start user sync job", zap.Error(err))
		os.Exit(1)
	}

	cron.Start()
	_ = ants.Submit(users.ListenUserEvents)         // 监听用户上下线
	_ = ants.Submit(traffic.ListenEventConsumer)    // 监听mmtls
	_ = ants.Submit(traffic.ListenSNIEventConsumer) // 监听sni

	if config.Debug {
		_ = ants.Submit(func() {
			log.Println(http.ListenAndServe(":6060", nil))
		})
	}
}

// 捕获任务完成后的处理
func handleCaptureCompletion(cancel context.CancelFunc, assembly *analyze.Analyze) {
	cancel()
	closed := assembly.Assembler.FlushAll()
	assembly.Factory.WaitGoRoutines()

	zap.L().Info("Flushed stream", zap.Int("count", closed))
}

// 优雅退出
func handleGracefulExit(cancel context.CancelFunc) {
	cancel()

	spinners.Start(i18n.T("handle graceful wait exit"))
	// 停止 cron，超时退出避免阻塞
	done := make(chan struct{})
	go func() {
		cron.Stop()
		close(done)
	}()

	select {
	case <-done:
		zap.L().Info("Cron stopped")
	case <-time.After(2 * time.Second):
		zap.L().Warn("Cron stop timeout")
	}

	// 释放协程池
	ants.Release()
	zap.L().Info("Release goroutine pool")

	// 刷新日志并退出
	_ = zap.L().Sync()
	time.Sleep(500 * time.Millisecond)
	spinners.Stop("capture stop", nil)
	os.Exit(0)
}
