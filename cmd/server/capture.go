package server

import (
	"context"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/socket/handler"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/full"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/keywords"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/brands/partial"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
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

func NewCaptureServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "capture",
		Short:  "capture commands",
		PreRun: captureRreFunc,
		Run:    captureCmdRun,
	}
	cmd.Flags().StringVar(&config.CaptureNic, "nic", config.Cfg.Capture.NIC, "capture nic")
	cmd.Flags().StringVar(&config.CapturePcap, "pcap", config.Cfg.Capture.OfflineFile, "capture pcap file")
	cmd.Flags().BoolVar(&config.UseMongo, "use-mongo", config.Cfg.UseMongo, "use mongo db")
	cmd.Flags().BoolVar(&config.UseFeature, "feature", config.Cfg.UseFeature, "use parse application")
	cmd.Flags().StringVar(&config.BerkeleyPacketFilter, "bpf", config.Cfg.BerkeleyPacketFilter, "Berkeley packet filter")
	cmd.Flags().BoolVar(&config.IgnoreMissing, "ignore-missing", config.Cfg.IgnoreMissing, "ignore missing packet")
	cmd.Flags().BoolVar(&config.FollowOnlyOnlineUsers, "follow-online-users", config.Cfg.FollowOnlyOnlineUsers, "follow only online users")
	cmd.Flags().BoolVar(&config.UseTTL, "use-ttl", config.Cfg.UseTTL, "save TTL for IP")
	cmd.Flags().BoolVar(&config.UseUA, "use-ua", config.Cfg.UseUA, "use ua parser")
	cmd.Flags().StringVar(&config.Geo2IP, "geo2ip", config.Cfg.Geo2IP, "geo2ip")
	cmd.Flags().BoolVarP(&config.Detach, "detach", "d", config.Cfg.Detach, "Run server in background and print PID")
	cmd.Flags().StringVar(&config.UaRegular, "ua_regular", config.Cfg.UaRegular, "UserAgent Regular YAML File")
	return cmd
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

// TODO 流量来源与目的地 源IP和目标IP热图（MaxMind GeoIP）、最频繁访问目标IP

// captureRreFunc 捕获前置方法
func captureRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})
}

// captureCmdRun 捕获子命令入口
func captureCmdRun(cmd *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	switch cmd.Parent().Name() {
	case "stop":
		captureDaemon.Stop()
		break
	case "status":
		captureDaemon.Status()
		break
	case "restart":
		captureDaemon.Restart(webRun)
		break
	default:
		if config.Detach {
			captureDaemon.Start(captureRun)
			break
		}
		captureRun()
		break
	}
	return
}

// 捕获开始
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

	if config.UseFeature {
		if err = features.Setup(); err != nil {
			os.Exit(1)
		}
	}

	if config.UseUA {
		if err = uaparser.Setup(); err != nil {
			os.Exit(1)
		}
	}

	//if config.Geo2IP != "" {
	//	maxmind.MaxMind.Filename = fmt.Sprintf("%s/%s", config.EtcDir, config.Geo2IP)
	//	if err = maxmind.MaxMind.Setup(); err != nil {
	//		//os.Exit(1)
	//	}
	//}
	// 加载品牌精确匹配
	if err = full.Brands.Setup(); err != nil {
		os.Exit(1)
	}
	// 加载品牌部分匹配
	if err = partial.Brands.Setup(); err != nil {
		os.Exit(1)
	}
	// 加载品牌关键词匹配
	if err = keywords.Brands.Setup(); err != nil {
		os.Exit(1)
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
