package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/logger"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"os"
)

const (
	CliName    = "dpi"
	CliVersion = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:    CliName,
	Short:  CliVersion,
	PreRun: PreFunc,
	Run:    RunFunc,
	PersistentPreRun: func(c *cobra.Command, args []string) {
		// 加载日志组件
		logger := &logger.Logger{}
		_ = logger.Setup()
		// 加载翻译组件
		err := i18n.I18n.Setup()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	// define flags
	rootCmd.PersistentFlags().StringVarP(&config.Language, "language", "l", config.Cfg.Language, "language")
	rootCmd.PersistentFlags().StringVar(&config.LogLevel, "log-level", config.Cfg.LogLevel, "log level")
	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "D", config.Cfg.Debug, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&config.Signal, "signal", "s", "", "send signal to a master process: stop, quit, reopen, reload")

	// define sub command
	rootCmd.AddCommand(CaptureCmd)
	rootCmd.AddCommand(WebCmd)
}

func RunFunc(c *cobra.Command, args []string) {

}

func PreFunc(c *cobra.Command, args []string) {
	// 初始化翻译
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})
	if len(args) == 0 || c.Flags().NFlag() == 0 {
		_ = c.Help()
		os.Exit(0)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
