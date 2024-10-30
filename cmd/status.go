package cmd

import (
	"github.com/dot-xiaoyuan/dpi-analyze/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service running status",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	StatusCmd.AddCommand(server.NewCaptureServer())
	StatusCmd.AddCommand(server.NewWebServer())
}
