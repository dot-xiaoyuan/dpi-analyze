package traffic

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type Event struct {
	SourceIP string
	TargetIP string
	URL      string
}

var (
	eventChannel    = make(chan Event, 100)
	connectionCache sync.Map // 记录 （源IP + URL） -> ConnectionInfo
)

type ConnectionInfo struct {
	URL      string
	TargetIP map[string]struct{}
}

func SendMMTLSEvent(sourceIP, targetIP, url string) {
	event := Event{SourceIP: sourceIP, TargetIP: targetIP, URL: url}
	eventChannel <- event
}

func handleEvent(event Event) {
	key := fmt.Sprintf("%s-%s", event.SourceIP, event.URL)

	// 检查缓存中是否已存在该 (源IP + URL) 组合
	value, ok := connectionCache.Load(key)
	if !ok {
		// 不存在
		info := ConnectionInfo{URL: event.URL, TargetIP: make(map[string]struct{})}
		info.TargetIP[event.TargetIP] = struct{}{}
		connectionCache.Store(key, info)
	} else {
		// 已存在
		info := value.(ConnectionInfo)
		if _, found := info.TargetIP[event.TargetIP]; !found {
			// 新目标IP
			info.TargetIP[event.TargetIP] = struct{}{}
			connectionCache.Store(key, info)
		}
	}
}

func ListenMMTLSEvent() {
	for event := range eventChannel {
		_ = ants.Submit(func() {
			handleEvent(event)
		})
	}
}

type MMTLSEvent struct {
	SourceIP string
	TargetIP string
	URL      string
}

var eventPool = sync.Pool{
	New: func() interface{} {
		return &MMTLSEvent{}
	},
}

// SendEvent2Redis 发送事件到redis队列
func SendEvent2Redis(s, d, u string) {
	_, err := redis.GetRedisClient().LPush(context.Background(),
		types.ListEventQueue,
		fmt.Sprintf("%s|%s|%s", s, d, u)).Result()
	if err != nil {
		zap.L().Error("Failed to push event 2 redis:", zap.Error(err))
	}
}

// 更新member
func processEvent(event MMTLSEvent) {
	key := fmt.Sprintf("connections:%s:%s", event.SourceIP, event.URL)

	_, err := redis.GetRedisClient().SAdd(context.Background(), key, event.TargetIP).Result()
	if err != nil {
		zap.L().Error("Failed to push event 2 redis:", zap.Error(err))
		return
	}

	count, _ := redis.GetRedisClient().SCard(context.Background(), key).Result()
	if count > 1 {
		zap.L().Debug("Multiple target IPs", zap.String("src_ip", event.SourceIP), zap.String("dst_ip", event.TargetIP))
		fmt.Printf("Multiple target IPs: %s\n", event.SourceIP)
	}
	redis.GetRedisClient().Expire(context.Background(), key, time.Minute*5).Val()
}

func ListenEventConsumer() {
	for {
		result, err := redis.GetRedisClient().BRPop(context.Background(), 0, types.ListEventQueue).Result()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		data := result[1]
		parts := strings.Split(data, "|")
		if len(parts) != 3 {
			zap.L().Error("Invalid event format", zap.String("data", data))
			continue
		}

		event := eventPool.Get().(*MMTLSEvent)
		event.SourceIP = parts[0]
		event.TargetIP = parts[1]
		event.URL = parts[2]
		_ = ants.Submit(func() {
			defer eventPool.Put(event)
			processEvent(*event)
		})
	}
}
