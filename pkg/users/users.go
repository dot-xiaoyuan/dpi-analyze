package users

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

var (
	OnlineUsers sync.Map
	hashFields  = []string{"rad_online_id", "user_name", "ip", "user_mac", "line_type", "add_time", "products_id", "billing_id", "control_id"}
)

type UserEvent types.UserEvent

func getHash(id string, rdb *v9.Client) types.User {
	hashKey := fmt.Sprintf(types.HashRadOnline, id)
	var user types.User
	if err := rdb.HMGet(context.TODO(), hashKey, hashFields...).Scan(&user); err != nil {
		zap.L().Error("Error getting hash", zap.String("hashKey", hashKey), zap.Error(err))
		return types.User{}
	}

	return user
}

// 记录用户，内部用sync.map。有序列表使用z set
func storeUser(ip string, user types.User) {
	OnlineUsers.LoadOrStore(ip, user)
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZAdd(ctx, types.ZSetOnlineUsers, v9.Z{
		Score:  float64(user.AddTime),
		Member: ip,
	}).Val()
	rdb.HMSet(ctx, fmt.Sprintf(types.HashAnalyzeIP, ip), "username", user.UserName, "mac", user.UserMac).Val()
}

// DropUser 记录用户，下线删除在线表中的IP
func DropUser(ip string) {
	OnlineUsers.Delete(ip)
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZRem(ctx, types.ZSetIP, ip).Val()
	rdb.ZRem(ctx, types.ZSetOnlineUsers, ip).Val()
	rdb.Del(ctx, fmt.Sprintf(types.HashAnalyzeIP, ip)).Val()
	rdb.Del(ctx, fmt.Sprintf(types.SetIPDevices, ip)).Val()

	member.DelMemory(ip)
	member.DelFeatureSet(ip)
	observer.TTLObserver.DeleteRedis(ip)
	observer.MacObserver.DeleteRedis(ip)
	observer.UaObserver.DeleteRedis(ip)
	observer.DeviceObserver.DeleteRedis(ip)
}

// FindUser 查找用户
func FindUser(ip string) types.User {
	user, ok := OnlineUsers.Load(ip)
	if ok {
		return user.(types.User)
	}
	return types.User{}
}

// FindUserName 根据ip查找用户名
func FindUserName(ip string) string {
	user, ok := OnlineUsers.Load(ip)
	if ok {
		return user.(types.User).UserName
	}
	return ""
}

// ExitsUser 用户是否存在
func ExitsUser(ip string) bool {
	_, ok := OnlineUsers.Load(ip)
	return ok
}

func Traversal(c types.Condition) (int64, interface{}, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (c.Page - 1) * c.PageSize

	zap.L().Debug("Online Users 偏移量", zap.Int64("start", start), zap.Int64("page", c.Page), zap.Int64("size", c.PageSize))
	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	count := rdb.ZCount(ctx, types.ZSetOnlineUsers, c.Min, c.Max).Val()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, types.ZSetOnlineUsers, &v9.ZRangeBy{
		Min:    c.Min,      // 查询范围的最小时间戳
		Max:    c.Max,      // 查询范围的最大时间戳
		Offset: start,      // 分页起始位置
		Count:  c.PageSize, // 每页大小
	})

	_, err := pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("ZRange pipe.Exec", zap.Error(err))
		return 0, nil, err
	}

	ips := zRangCmd.Val()
	var result []types.User
	for _, ip := range ips {
		user := FindUser(ip.Member.(string))
		result = append(result, user)
	}
	return count, result, nil
}

// LoadEvent 上线事件
// 1.更新在线表
// 2.记录事件日志2mongo
func (u *UserEvent) LoadEvent() {
	storeUser(u.Ip, types.User{
		UserName:   u.UserName,
		IP:         u.Ip,
		UserMac:    u.UserMac,
		LineType:   u.LineType,
		AddTime:    u.AddTime,
		ProductsID: u.ProductsId,
		BillingID:  u.BillingId,
		ContractID: u.ControlId,
	})
	u.Save2Mongo()
}

// DropEvent 下线事件
func (u *UserEvent) DropEvent() {
	DropUser(u.Ip)
	u.Save2Mongo()
}

func (u *UserEvent) Save2Mongo() {
	ctx := context.TODO()

	_, err := mongodb.GetMongoClient().Database(types.MongoDatabaseUserEvents).Collection(time.Now().Format("06_01")).InsertOne(ctx, u)
	if err != nil {
		zap.L().Error(i18n.T("Error inserting event"), zap.Error(err))
		os.Exit(1)
	}
}
