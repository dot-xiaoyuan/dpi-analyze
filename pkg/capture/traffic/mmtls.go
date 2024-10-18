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
		types.ListMMTLSEventQueue,
		fmt.Sprintf("%s|%s|%s", s, d, u)).Result()
	if err != nil {
		zap.L().Error("Failed to push event 2 redis:", zap.Error(err))
	}
}

// 更新member
func processEvent(event MMTLSEvent) {
	key := fmt.Sprintf(types.SetMMTLSConnection, event.SourceIP, event.URL)

	_, err := redis.GetRedisClient().SAdd(context.Background(), key, event.TargetIP).Result()
	if err != nil {
		zap.L().Error("Failed to push event 2 redis:", zap.Error(err))
		return
	}

	count, _ := redis.GetRedisClient().SCard(context.Background(), key).Result()
	if count >= 2 {
		zap.L().Debug("Multiple target IPs", zap.String("src_ip", event.SourceIP), zap.String("dst_ip", event.TargetIP))
		fmt.Printf("Multiple target IPs: %s, member => %s\n", event.SourceIP, key)
		dips := redis.GetRedisClient().SMembers(context.Background(), key).Val()
		for _, dip := range dips {
			fmt.Printf("dips:%s\n", dip)
		}
	}
	redis.GetRedisClient().Expire(context.Background(), key, time.Minute*5).Val()
}

func ListenEventConsumer() {
	for {
		result, err := redis.GetRedisClient().BRPop(context.Background(), 0, types.ListMMTLSEventQueue).Result()
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
