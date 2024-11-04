package cmd

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// ps

var PsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List DPI Server",
	Run: func(cmd *cobra.Command, args []string) {
		var processes [][]string

		// 读取 run 目录中的文件
		entries, err := os.ReadDir(config.RunDir)
		if err != nil {
			fmt.Println("Error reading run directory:", err)
			return
		}

		// 遍历每个文件，读取 PID
		for _, entry := range entries {
			if !entry.IsDir() {
				split := strings.Split(entry.Name(), ".")
				if split[len(split)-1] != "pid" {
					continue
				}
				// 读取 PID 文件
				pidFile := filepath.Join(config.RunDir, entry.Name())
				pidBytes, err := os.ReadFile(pidFile)
				if err != nil {
					fmt.Printf("Error reading PID file %s: %v\n", pidFile, err)
					continue
				}

				// 转换为整数
				pidStr := strings.TrimSpace(string(pidBytes))
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					fmt.Printf("Invalid PID in file %s: %s\n", pidFile, pidStr)
					continue
				}

				// 获取进程信息并添加到 processes 切片
				processInfo := getProcessInfo(pid)
				if processInfo != nil {
					processes = append(processes, processInfo)
				}
			}
		}

		// 显示进程信息表格
		displayProcesses(processes)
	},
}

func getProcessInfo(pid int) []string {
	var cmd *exec.Cmd

	// 根据操作系统选择命令
	if runtime.GOOS == "linux" {
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,comm,args,state,user,time,%mem,%cpu")
	} else if runtime.GOOS == "darwin" { // macOS
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,comm,args,state,user,time,pmem,pcpu")
	} else {
		fmt.Println("Unsupported OS")
		return nil
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting info for PID %d: %v\n", pid, err)
		return nil
	}

	// 处理输出，返回信息切片
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil // 没有有效的信息
	}

	// 分割输出信息并将 args 合并为一个字符串
	fields := strings.Fields(lines[1])
	if len(fields) > 7 { // 确保 args 存在
		args := strings.Join(fields[2:len(fields)-5], " ") // 合并 args 列，但不包括最后 5 列
		prefix := append(fields[:2], args)                 // 将合并后的 args 放回切片
		fields = append(prefix, fields[len(fields)-5:]...) // 添加剩下的列
	}

	return fields
}

func displayProcesses(processes [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PID", "Command", "Args", "State", "User", "Time", "Memory", "CPU"})

	for _, process := range processes {
		table.Append(process)
	}

	table.Render() // 打印表格
}
