package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/statistics"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"log"
	"os"
)

// 统计cli

var StatisticsCmd = &cobra.Command{
	Use:    "statistics",
	Short:  "statistics commands",
	PreRun: StatisticsRreFunc,
	Run:    StatisticsRun,
}

func init() {
	// define flag
	StatisticsCmd.Flags().UintVar(&config.StatisticsPort, "port", 8088, "statistics port to listen on")
}

func StatisticsRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.T(c.Short)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.T(flag.Usage)
	})

	spinners.Setup()
	zap.L().Info(i18n.T("Start Load Mongodb Component"))
	mongo.Setup()
}

func StatisticsRun(c *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	statistics.StartStatistics(statistics.Config{
		Port: config.StatisticsPort,
	})
}
