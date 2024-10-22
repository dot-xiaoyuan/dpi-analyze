package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"os"
)

// 统计cli

var WebCmd = &cobra.Command{
	Use:    "web",
	Short:  "web commands",
	PreRun: WebRreFunc,
	Run:    WebRun,
}

var webDaemon = &utils.Daemon{
	Name: "web",
	Context: daemon.Context{
		PidFileName: fmt.Sprintf("%s/web.pid", config.RunDir),
		PidFilePerm: 0644,
		LogFileName: fmt.Sprintf("%s/web.log", config.LogDir),
		LogFilePerm: 0640,
		WorkDir:     config.Home,
		Args:        os.Args,
		Umask:       027,
	},
}

func init() {
	// define flag
	WebCmd.Flags().UintVar(&config.WebPort, "port", 8088, "web port to listen on")
	WebCmd.Flags().BoolVarP(&config.Detach, "detach", "d", config.Cfg.Detach, "Run web in background and print process ID")
}

func WebRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})
}

func WebRun(*cobra.Command, []string) {
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
			webDaemon.Stop()
		case types.START:
			webDaemon.Start(webRun)
		case types.STATUS:
			webDaemon.Status()
		case types.RESTART:
			webDaemon.Restart(webRun)
		default:
			fmt.Println("Usage: [start|stop|status|restart]")
		}
		os.Exit(0)
	}
	if config.Detach {
		webDaemon.Start(webRun)
		return
	}
	webRun()
}

func webRun() {
	web.NewWebServer(web.Config{Port: config.WebPort})
}
