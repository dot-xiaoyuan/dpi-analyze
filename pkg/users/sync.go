package users

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	types2 "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"os"
	"time"
)

// 用户同步

type UserSync struct{}

// CleanUp 清空用户在线表
func (us UserSync) CleanUp() {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.Del(ctx, types2.ZSetOnlineUsers).Val()
}

// Run SyncOnlineUsers 同步在线用户
func (us UserSync) Run() {
	err := SyncOnlineUsers()
	if err != nil {
		os.Exit(1)
	}
}

func SyncOnlineUsers() error {
	rdb := redis.GetOnlineRedisClient()
	ctx := context.TODO()

	ids, err := rdb.LRange(ctx, types2.ListRadOnline, 0, -1).Result()
	if err != nil {
		zap.L().Error(i18n.T("SyncOnlineUsers error"), zap.Error(err))
		return err
	}

	var count int
	for _, id := range ids {
		user := getHash(id, rdb)
		if user.UserName != "" {
			count++
			storeUser(user.IP, user)
		}
	}
	zap.L().Info(i18n.T("SyncOnlineUsers"), zap.Int("count", count))
	return nil
}

func ListenUserEvents() {
	rdb := redis.GetCacheRedisClient()
	ctx := context.TODO()

	listKey := fmt.Sprintf(types2.ListAntiProxy, config.Cfg.Redis.DPI.Host)

	for {
		event, err := rdb.BLPop(ctx, time.Minute, listKey).Result()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		userEvent := UserEvent{}
		_ = json.Unmarshal([]byte(event[1]), &userEvent)
		zap.L().Info(i18n.T("Listen user events"), zap.Int("action", userEvent.Action), zap.String("username", userEvent.UserName))

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
