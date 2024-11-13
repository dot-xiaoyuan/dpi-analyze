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
	CliName     = "dpi-analyze"
	Description = "dpi CLI"
	CliVersion  = "1.0.4.241113_beta"
)

var rootCmd = &cobra.Command{
	Use:     CliName,
	Short:   Description,
	PreRun:  rootPreFunc,
	Run:     rootRunFunc,
	Version: CliVersion,
	PersistentPreRun: func(c *cobra.Command, args []string) {
		// 加载日志组件
		logger := &logger.Logger{}
		_ = logger.Setup()
		// 加载翻译组件
		i18n.I18n.Lang = config.Language
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

	// Global Flags
	rootCmd.PersistentFlags().StringVar(&config.Language, "lang", config.Cfg.Language, "language")
	rootCmd.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "l", config.Cfg.LogLevel, "Set the logging level ('debug', 'info', 'warn', 'error', 'fatal')")
	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "D", config.Cfg.Debug, "Enable debug mode")

	rootCmd.AddCommand(RunCmd)
	rootCmd.AddCommand(PsCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(StatusCmd)
	rootCmd.AddCommand(RestartCmd)
}

func rootRunFunc(c *cobra.Command, args []string) {

}

func rootPreFunc(c *cobra.Command, args []string) {
	// 初始化翻译
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})
	if len(args) == 0 || c.Flags().NFlag() == 0 {
		_ = c.Help()
		os.Exit(0)
	}
}

// 自定义 Help 模板
var customHelpTemplate = `Usage:
  {{.UseLine}}

{{if .Commands}}Common Commands:
{{range .Commands}}{{.Name | printf "  %-11s"}} {{.Short}}
{{end}}{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
