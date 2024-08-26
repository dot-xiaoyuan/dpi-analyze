package analyze

import "C"
import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/config"
	"github.com/spf13/cobra"
)

var CaptureCmd = &cobra.Command{
	Use: "capture",
	PreRun: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = c.Help()
			return
		}
	},
	Run: CaptureCmdRun,
}

func CaptureCmdRun(c *cobra.Command, args []string) {
}

func init() {
	initFlags()
}

func initFlags() {
	CaptureCmd.PersistentFlags().StringVar(&config.CaptureNic, "nic", config.C.NIC, "Capture nic")
	CaptureCmd.PersistentFlags().StringVar(&config.CapturePcap, "pcap", config.C.OfflineFile, "Capture pcap file")
	CaptureCmd.PersistentFlags().BoolVarP(&config.UseMongo, "mongo", "m", config.C.UseMongo, "UseMongo Mongodb")
}
