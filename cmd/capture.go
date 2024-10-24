package cmd

import (
	"context"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/socket/handler"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/maxmind"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/uaparser"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	v3 "github.com/robfig/cron/v3"
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

var sp = spinner.New(spinner.CharSets[36], 100*time.Millisecond)
var cron = v3.New()

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
	defer cancel()

	// 启动系统信号监听
	go handleSignals(cancel)

	// 使用spinner加载组件
	loadComponents()

	// 启动 Packet Capture
	assembly := analyze.NewAnalyzer()
	done := make(chan struct{})
	defer close(done)

	go capture.StartCapture(ctx, capture.Config{
		OffLine:              config.CapturePcap,
		Nic:                  config.CaptureNic,
		SnapLen:              16 << 10,
		BerkeleyPacketFilter: config.BerkeleyPacketFilter,
	}, assembly, done)

	// 阻塞等待信号或完成
	// 等待捕获完成
	select {
	case <-done:
		handleCaptureCompletion(cancel, assembly)
		fmt.Println("capture completed")
		break
	case <-ctx.Done():
		fmt.Println("capture cancelled")
		break
	}
}

// 捕获任务完成后的处理
func handleCaptureCompletion(cancel context.CancelFunc, assembly *analyze.Analyze) {
	cancel()
	closed := assembly.Assembler.FlushAll()
	assembly.Factory.WaitGoRoutines()

	sp.Start()
	time.Sleep(time.Second * 3)
	zap.L().Info("Flushed stream", zap.Int("count", closed))
	sp.Stop()
}

// 信号监听 Goroutine
func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		zap.L().Info("Received signal", zap.String("signal", sig.String()))
		cancel() // 收到信号后立即取消上下文
		handleGracefulExit()
	}
}

// 优雅退出
func handleGracefulExit() {
	sp.Start()
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
	sp.Stop()
	os.Exit(0)
}

// 加载所有组件并使用 Spinner 提示
func loadComponents() {
	var err error
	if err = redis.Setup(); err != nil {
		os.Exit(1)
	}
	if err = ants.Setup(10); err != nil {
		os.Exit(1)
	}

	if config.UseMongo {
		if err = mongo.Setup(); err != nil {
			os.Exit(1)
		}
	}

	if config.ParseFeature {
		if err = features.Setup(); err != nil {
			os.Exit(1)
		}
	}

	if config.UseUA {
		if err = uaparser.Setup(); err != nil {
			os.Exit(1)
		}
	}

	if config.Geo2IP != "" {
		maxmind.MaxMind.Filename = config.Geo2IP
		if err = maxmind.MaxMind.Setup(); err != nil {
			os.Exit(1)
		}
	}
	// 注册unix路由
	handler.InitHandlers()

	// 在线用户同步组件
	// 1.运行后先清除遗留数据
	// 2.首次加载先全量加载一次，然后定时同步
	userSync := users.UserSync{}
	userSync.CleanUp()
	if err = users.SyncOnlineUsers(); err != nil {
		//sp.Stop()
		os.Exit(1)
	}

	_, err = cron.AddJob("@every 30m", userSync)
	if err != nil {
		zap.L().Error("Failed to start user sync job", zap.Error(err))
		os.Exit(1)
	}
	cron.Start()

	go socket.StartServer()
	go users.ListenUserEvents() // 监听用户上下线
	//_ = ants.Submit(traffic.ListenEventConsumer)    // 监听mmtls
	//_ = ants.Submit(traffic.ListenSNIEventConsumer) // 监听sni

	if config.Debug {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
}
