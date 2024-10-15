package utils

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
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
		return false, 0
	}
	pid, _ := strconv.Atoi(string(pidData))
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil, pid
}

func (daemon *Daemon) Start(run func()) {
	running, pid := daemon.isRunning()
	if running {
		fmt.Printf("%s server already running with %d.\n", daemon.Name, pid)
		return
	}
	d, err := daemon.Context.Reborn()
	if err != nil {
		fmt.Printf("Unable to run: %s\n", err)
		os.Exit(1)
	}
	if d != nil {
		fmt.Printf("%s server start successfully with %d.\n", daemon.Name, pid)
		os.Exit(0)
	}
	defer daemon.Context.Release()
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
	spinners.Start()
	defer spinners.Stop()
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
