package users

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
)

func HookDropUser(user types.User, pr *types.ProxyRecord) error {
	// 这里不再次查询在线表了，直接写
	rdb := redis.GetOnlineRedisClient()
	ctx := context.TODO()

	onlineKey := fmt.Sprintf(types.HashRadOnline, user.RadOnlineID)
	if ok := rdb.Exists(ctx, onlineKey).Val(); ok == 0 {
		zap.L().Error("在线信息不存在", zap.String("key", onlineKey))
		// 清空程序里该用户信息
		DropUser(user.IP)
		return nil
	}
	// 修改用户在线表 设备数
	rdb.HMSet(ctx, onlineKey, "pc_num", pr.PcCount, "mobile_num", pr.MobileCount).Val()

	cache := redis.GetCacheRedisClient()
	update := fmt.Sprintf("1-%d", user.RadOnlineID)

	zap.L().Info("模拟 UPDATE 报文推送", zap.String("队列", types.ListRadOnlineUpdate), zap.String("元素", update))
	_, err := cache.RPush(ctx, types.ListRadOnline, update).Result()
	if err != nil {
		zap.L().Error("推送虚拟下线报文失败", zap.Error(err))
		return err
	}
	return nil
}
