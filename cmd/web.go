package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
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

func init() {
	// define flag
	WebCmd.Flags().UintVar(&config.WebPort, "port", 8088, "web port to listen on")
	WebCmd.Flags().BoolVarP(&config.Detach, "detach", "d", config.Cfg.Detach, "Run web in background and print process ID")
}

func WebRreFunc(c *cobra.Command, args []string) {
	contxt := &daemon.Context{
		PidFileName: "web.pid",
		PidFilePerm: 0644,
		LogFileName: "web.log",
		LogFilePerm: 0640,
		WorkDir:     fmt.Sprintf("%s/run/", config.Home),
		Args:        []string{"[go-daemon sample]"},
		Umask:       027,
	}

	d, err := contxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer contxt.Release()

	log.Print("-----------------")
	log.Print("daemon started")

	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})

	spinners.Setup()
	zap.L().Info(i18n.T("Start Load Mongodb Component"))
	mongo.Setup()
}

func WebRun(c *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	web.NewWebServer(web.Config{Port: config.WebPort})
}
