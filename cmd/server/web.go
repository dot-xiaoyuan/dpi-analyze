package server

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"os"
)

// 统计cli

func NewWebServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "web",
		Short:  "web commands",
		PreRun: webRreFunc,
		Run:    webCmdRun,
	}
	cmd.Flags().UintVar(&config.WebPort, "port", 8088, "web port to listen on")
	cmd.Flags().BoolVarP(&config.Detach, "detach", "d", config.Cfg.Detach, "Run server in background and print PID")
	return cmd
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

func webRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})
}

func webCmdRun(cmd *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	switch cmd.Parent().Name() {
	case "stop":
		webDaemon.Stop()
		break
	case "status":
		webDaemon.Status()
		break
	case "restart":
		webDaemon.Restart(webRun)
		break
	default:
		if config.Detach {
			webDaemon.Start(webRun)
			return
		}
		webRun()
	}
}

func webRun() {
	web.NewWebServer(web.Config{Port: config.WebPort})
}
