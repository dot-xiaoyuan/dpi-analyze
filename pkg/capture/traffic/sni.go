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

// SNIEvent 用于存储 SNI 事件
type SNIEvent struct {
	SourceIP string
	TargetIP string
	URL      string
}

var sniEventPool = sync.Pool{
	New: func() interface{} {
		return &SNIEvent{}
	},
}

// SendSNIEvent2Redis 发送 SNI 事件到 Redis 队列
func SendSNIEvent2Redis(s, d, u string) {
	_, err := redis.GetRedisClient().LPush(context.Background(),
		types.ListSniEventQueue,
		fmt.Sprintf("%s|%s|%s", s, d, u)).Result()
	if err != nil {
		zap.L().Error("Failed to push SNI event to Redis:", zap.Error(err))
	}
}

// 更新 SNI member
func processSNIEvent(event SNIEvent) {
	key := fmt.Sprintf(types.SetSNIConnection, event.SourceIP, event.URL)

	fmt.Printf("processSNIEvent: [%s] => [%s]\n", key, event.URL)
	_, err := redis.GetRedisClient().SAdd(context.Background(), key, event.TargetIP).Result()
	if err != nil {
		zap.L().Error("Failed to add SNI event to Redis:", zap.Error(err))
		return
	}

	count, _ := redis.GetRedisClient().SCard(context.Background(), key).Result()
	if count > 1 {
		zap.L().Debug("Multiple target IPs for SNI", zap.String("src_ip", event.SourceIP), zap.String("dst_ip", event.TargetIP))
		fmt.Printf("Multiple target IPs for SNI: %s, member => %s\n", event.SourceIP, key)
	}
	redis.GetRedisClient().Expire(context.Background(), key, time.Minute*5).Val()
}

// ListenSNIEventConsumer 监听 SNI 事件消费者
func ListenSNIEventConsumer() {
	for {
		result, err := redis.GetRedisClient().BRPop(context.Background(), 0, types.ListSniEventQueue).Result()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		data := result[1]
		parts := strings.Split(data, "|")
		if len(parts) != 3 {
			zap.L().Error("Invalid SNI event format", zap.String("data", data))
			continue
		}

		event := sniEventPool.Get().(*SNIEvent)
		event.SourceIP = parts[0]
		event.TargetIP = parts[1]
		event.URL = parts[2]
		_ = ants.Submit(func() {
			defer sniEventPool.Put(event)
			processSNIEvent(*event)
		})
	}
}
