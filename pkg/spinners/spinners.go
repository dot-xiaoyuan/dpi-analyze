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
	mu      sync.Mutex // 防止并发冲突
	count   int
)

// Setup 初始化 Spinner 实例，只执行一次
func Setup() {
	once.Do(func() {
		loadSpinner()
	})
}

// 创建 Spinner 实例
func loadSpinner() {
	Spinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
}

// Start 启动加载动画并附加任务描述
func Start(task string) {
	mu.Lock()
	defer mu.Unlock()

	count++
	formattedTask := formatTask(fmt.Sprintf("[%d] %s...", count, task), 50)
	Spinner.Prefix = formattedTask
	_ = Spinner.Color("green", "bold")

	Spinner.Start()
}

// Stop 停止加载动画并打印状态信息
func Stop(task string, err error) {
	mu.Lock()
	defer mu.Unlock()

	Spinner.Stop()

	// 清除加载动画残留的行，避免显示错乱
	_, _ = fmt.Fprintf(os.Stderr, "\r\033[K")

	formattedTask := formatTask(fmt.Sprintf("%s", task), 50)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s Error: %s\n", formattedTask, err.Error())
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s Done\n", formattedTask)
}

// WithSpinner 通过加载组件调用函数并处理错误
func WithSpinner(task string, fn func() error) {
	Start(task)

	if err := fn(); err != nil {
		Stop(task, err) // 有错误时停止并打印
		return
	}

	Stop(task, nil) // 成功完成
}

// 格式化任务字符串
func formatTask(task string, width int) string {
	return fmt.Sprintf("%-*s", width, task)
}
