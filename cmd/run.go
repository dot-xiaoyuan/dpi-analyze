package cmd

import (
	"github.com/dot-xiaoyuan/dpi-analyze/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

// 运行子命令

var RunCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Create and run server",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	RunCmd.AddCommand(server.NewCaptureServer())
	RunCmd.AddCommand(server.NewWebCmd())
}
