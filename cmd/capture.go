package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var CaptureCmd = &cobra.Command{
	Use:   "capture",
	Short: "capture commands",
	Run:   CaptureRun,
}

func CaptureRun(c *cobra.Command, args []string) {
	fmt.Println("capture called")
}
