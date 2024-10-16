package utils

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"os"
	"strconv"
	"syscall"
	"time"
)

type Daemon struct {
	Name    string
	Context daemon.Context
}

type daemonProcess interface {
	isRunning() (bool, int)
	Start(run func())
	Stop()
	Restart(run func())
	Status()
}

func (daemon *Daemon) isRunning() (bool, int) {
	pidData, err := os.ReadFile(daemon.Context.PidFileName)
	if err != nil {
		// 读取 PID 文件失败，表示进程未运行
		return false, 0
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil || pid <= 0 {
		// PID 文件内容无效
		return false, 0
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid
	}

	// 发送 0 信号检测进程是否存在
	if err = process.Signal(syscall.Signal(0)); err != nil {
		return false, pid
	}
	return true, pid
}

func (daemon *Daemon) Start(run func()) {
	running, pid := daemon.isRunning()
	if running {
		fmt.Printf("%s server already running with PID %d.\n", daemon.Name, pid)
		return
	}

	d, err := daemon.Context.Reborn()
	if err != nil {
		fmt.Printf("Unable to start daemon: %v\n", err)
		os.Exit(1)
	}

	if d != nil {
		// 子进程启动成功，但父进程结束
		fmt.Printf("%s server started successfully with PID %d.\n", daemon.Name, d.Pid)
		os.Exit(0)
	}

	// 在子进程中执行
	defer daemon.Context.Release()

	// 确保在启动后写入 PID 文件
	err = daemon.writePidFile()
	if err != nil {
		fmt.Printf("Failed to write PID file: %v\n", err)
		os.Exit(1)
	}

	// 启动主逻辑
	run()
}

func (daemon *Daemon) Stop() {
	running, pid := daemon.isRunning()
	if !running {
		fmt.Printf("%s server is not running.\n", daemon.Name)
		return
	}
	process, _ := os.FindProcess(pid)
	if err := process.Kill(); err != nil {
		fmt.Printf("Unable to kill process: %s.\n", err)
	} else {
		// 删除pid文件
		_ = os.Remove(daemon.Context.PidFileName)
		fmt.Printf("%s server stopped successfully.\n", daemon.Name)
	}
	return
}

func (daemon *Daemon) Restart(run func()) {
	daemon.Stop()
	time.Sleep(time.Second)
	daemon.Start(run)
	return
}

func (daemon *Daemon) Status() {
	running, pid := daemon.isRunning()
	if !running {
		fmt.Printf("%s server is not running.\n", daemon.Name)
	} else {
		fmt.Printf("%s server is running with %d.\n", daemon.Name, pid)
	}
	return
}

func (daemon *Daemon) writePidFile() error {
	pid := os.Getpid() // 获取当前进程 PID
	return os.WriteFile(daemon.Context.PidFileName, []byte(fmt.Sprintf("%d", pid)), 0644)
}
