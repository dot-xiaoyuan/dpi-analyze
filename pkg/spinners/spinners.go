package spinners

import (
	"fmt"
	"github.com/briandowns/spinner"
	"os"
	"sync"
	"time"
)

var (
	once    sync.Once
	Spinner *spinner.Spinner
	mu      sync.Mutex // 用于防止并发冲突
	count   int
)

// Setup 初始化 Spinner 实例，只执行一次
func Setup() {
	once.Do(func() {
		loadSpinner()
	})
}

// 内部函数：创建一个 Spinner 实例
func loadSpinner() {
	Spinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
}

// Start 启动加载动画并附加任务描述
func Start(task string) {
	count++
	mu.Lock()
	defer mu.Unlock()
	time.Sleep(1 * time.Second)
	formattedTask := formatTask(fmt.Sprintf("[%d] %s...", count, task), 50)
	Spinner.Prefix = formattedTask
	_ = Spinner.Color("green", "bold")
	Spinner.Start()
}

// Stop 停止加载动画并打印完成信息
func Stop(task string) {
	mu.Lock()
	defer mu.Unlock()
	Spinner.Stop()
	formattedTask := formatTask(fmt.Sprintf("%s", task), 50)
	_, _ = fmt.Fprintf(os.Stderr, "%s Done\n", formattedTask)
}

// WithSpinner 通过加载组件调用方法
func WithSpinner(task string, fn func()) {
	Start(task)
	defer Stop(task)
	fn()
}

func formatTask(task string, width int) string {
	return fmt.Sprintf("%-*s", width, task)
}
