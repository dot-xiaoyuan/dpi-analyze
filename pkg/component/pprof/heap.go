package pprof

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

const (
	Ki = 1024
	Mi = Ki * Ki
	Gi = Ki * Mi
	Ti = Ki * Gi
	Pi = Ki * Ti
)

type Mouse struct{ buffer [][Mi]byte }

func (m *Mouse) StealMem() {
	max := Gi
	for len(m.buffer)*Mi < max {
		m.buffer = append(m.buffer, [Mi]byte{})
	}
}

func CollectHeap() { // 设置采样率，默认每分配512*1024字节采样一次。如果设置为0则禁止采样，只能设置一次
	runtime.MemProfileRate = 512 * 1024
	f, err := os.Create("./heap.prof")
	if err != nil {
		log.Fatal("could not create heap profile: ", err)
	}
	defer f.Close()
	// 高的内存占用 : 有个循环会一直向 m.buffer 里追加长度为 1 MiB 的数组，直到总容量到达 1 GiB 为止，且一直不释放这些内存，这就难怪会有这么高的内存占用了。
	m := &Mouse{}
	m.StealMem()
	// runtime.GC() // 执行GC，避免垃圾对象干扰
	//将剖析概要信息记录到文件
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write heap profile: ", err)
	}
}
