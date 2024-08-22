package main

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/config"
	_ "github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:    "dpi",
	Short:  "dpi CLI command",
	PreRun: PreRunFunc,
	Run:    RunFunc,
}

func RunFunc(c *cobra.Command, args []string) {
	fmt.Println("language:", config.Language)
	fmt.Println("Run")
}

// PreRunFunc before run
func PreRunFunc(c *cobra.Command, args []string) {
	config.Setup()
	fmt.Println("PreRun")
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVarP(&config.Language, "language", "l", "en", "language")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main1() {
	//log.SetOutput(os.Stdout)
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Printf("PANIC : %v", err)
	//		fmt.Println("发生严重错误，请联系支持人员。")
	//		os.Exit(1)
	//	}
	//}()

	// 加载日志
	//logger.InitLogger(*config.LogLevel)

	// Make Context
	//ctx, cancel := context.WithCancel(context.Background())

	// Packet Capture
	//assembly := analyze.NewAnalyzer()
	//done := make(chan struct{})
	//go capture.StartCapture(ctx, capture.Config{
	//	OffLine: *config.CaptureOfflineFile,
	//	Nic:     *config.CaptureNic,
	//	SnapLen: *config.CaptureSnapLen,
	//}, assembly, done)

	// 信号阻塞
	//signalChan := make(chan os.Signal, 1)
	//signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	//
	//select {
	//case <-done:
	//	cancel()
	//	closed := assembly.Assembler.FlushAll()
	//	assembly.Factory.WaitGoRoutines()
	//	log.Printf("Flushed %d streams\n", closed)
	//case <-signalChan:
	//	cancel()
	//	log.Println("Received an interrupt, stopping analysis...")
	//	time.Sleep(time.Second)
	//	os.Exit(0)
	//}

}
