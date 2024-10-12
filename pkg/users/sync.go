package users

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	"go.uber.org/zap"
	"os"
	"time"
)

// 用户同步

type UserSync struct{}

// CleanUp 清空用户在线表
func (us UserSync) CleanUp() {
	rdb := redis.GetOnlineRedisClient()
	ctx := context.TODO()

	rdb.Del(ctx, types.ZSetOnlineUsers).Val()
}

// Run SyncOnlineUsers 同步在线用户
func (us UserSync) Run() {
	rdb := redis.GetOnlineRedisClient()
	ctx := context.TODO()

	ids, err := rdb.LRange(ctx, types.ListRadOnline, 0, -1).Result()
	if err != nil {
		zap.L().Error(i18n.T("SyncOnlineUsers error"), zap.Error(err))
		os.Exit(1)
	}

	for _, id := range ids {
		user := getHash(id, rdb)
		if user.UserName != "" {
			storeUser(user.IP, user)
		}
	}
}

func ListenUserEvents() {
	rdb := redis.GetCacheRedisClient()
	ctx := context.TODO()

	listKey := fmt.Sprintf(types.ListAntiProxy, config.Cfg.Redis.DPI.Host)

	for {
		event, err := rdb.BLPop(ctx, time.Minute, listKey).Result()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		userEvent := UserEvent{}
		_ = json.Unmarshal([]byte(event[1]), &userEvent)
		zap.L().Debug(i18n.T("Listen user events"), zap.Int("action", userEvent.Action), zap.String("username", userEvent.UserName))

		if userEvent.UserName == "" {
			zap.L().Warn("user event is empty", zap.Strings("event", event))
			continue
		}
		if userEvent.Action == 1 {
			// 上线
			userEvent.LoadEvent()
		} else if userEvent.Action == 2 {
			// 下线
			userEvent.DropEvent()
		}
		time.Sleep(time.Second)
	}
}
