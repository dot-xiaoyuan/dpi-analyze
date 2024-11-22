package cmd

import (
	"github.com/dot-xiaoyuan/dpi-analyze/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

var RestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart one or more running server",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	RestartCmd.AddCommand(server.NewCaptureServer())
	RestartCmd.AddCommand(server.NewWebCmd())
}
