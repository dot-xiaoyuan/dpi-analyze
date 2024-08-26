package cmd

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
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

func init() {
	// define flag
	CaptureCmd.Flags().StringVar(&config.CaptureNic, "nic", config.Cfg.NIC, "capture nic")
	CaptureCmd.Flags().StringVar(&config.CapturePcap, "pcap", config.Cfg.OfflineFile, "capture pcap file")
}

func CaptureRreFunc(c *cobra.Command, args []string) {
	c.Short = i18n.Translate.T(c.Short, nil)
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = i18n.Translate.T(flag.Usage, nil)
	})

	if len(args) == 0 && c.Flags().NFlag() == 0 {
		_ = c.Help()
		os.Exit(0)
	}
}

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
		OffLine: config.CapturePcap,
		Nic:     config.CaptureNic,
		SnapLen: 16 << 10,
	}, assembly, done)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-done:
		cancel()
		closed := assembly.Assembler.FlushAll()
		assembly.Factory.WaitGoRoutines()
		log.Printf("Flushed %d stream\n", closed)
	case <-signalChan:
		cancel()
		log.Println("Received terminate signal, stop analyze...")
		time.Sleep(time.Second)
		os.Exit(0)
	}
}
