package cmd

import (
	"github.com/dot-xiaoyuan/dpi-analyze/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop one or more running server",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	StopCmd.AddCommand(server.NewCaptureServer())
	StopCmd.AddCommand(server.NewWebServer())
}
