package users

import (
	"context"
	"fmt"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	v9 "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

var (
	OnlineUsers sync.Map
	hashFields  = []string{"user_name", "ip", "user_mac", "line_type", "add_time", "products_id", "billing_id", "control_id"}
)

type UserEvent types.UserEvent

func getHash(id string, rdb *v9.Client) types.User {
	hashKey := fmt.Sprintf(types.HashRadOnline, id)
	var user types.User
	if err := rdb.HMGet(context.TODO(), hashKey, hashFields...).Scan(&user); err != nil {
		zap.L().Error("Error getting hash", zap.String("hashKey", hashKey), zap.Error(err))
		return types.User{}
	}

	zap.L().Debug("get user", zap.Any("user", user))
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
}

// DropUser 记录用户，下线删除在线表中的IP
func DropUser(ip string) {
	OnlineUsers.Delete(ip)
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	rdb.ZRem(ctx, types.ZSetOnlineUsers, ip).Val()
}

// FindUser 查找用户
func FindUser(ip string) types.User {
	user, ok := OnlineUsers.Load(ip)
	if ok {
		return user.(types.User)
	}
	return types.User{}
}

// ExitsUser 用户是否存在
func ExitsUser(ip string) bool {
	_, ok := OnlineUsers.Load(ip)
	return ok
}

func Traversal(c provider.Condition) (int64, interface{}, error) {
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

	_, err := mongodb.GetMongoClient().Database("events").Collection(time.Now().Format("user_event_06_01")).InsertOne(ctx, u)
	if err != nil {
		zap.L().Error(i18n.T("Error inserting event"), zap.Error(err))
		os.Exit(1)
	}
}

func UserEventQuery(c provider.UserEventCondition) (int64, any, error) {
	zap.L().Info(i18n.T("UserEventQuery"), zap.Any("Condition", c))
	skip := (c.Page - 1) * c.PageSize
	// 构建公共的 match 条件
	matchStage := bson.D{
		{"$match", bson.D{}},
	}
	var sort int
	if c.OrderBy == "descend" {
		sort = -1
	} else {
		sort = 1
	}
	// 排序
	sortStage := bson.D{
		{"$sort", bson.D{{c.SortField, sort}}},
	}
	// 分页
	limitStage := bson.D{
		{"$limit", c.PageSize},
	}
	skipStage := bson.D{
		{"$skip", skip},
	}
	pipeline := mongo.Pipeline{matchStage, sortStage, limitStage, skipStage}

	coll := mongodb.GetMongoClient().
		Database("events").
		Collection(fmt.Sprintf("user_event_%s_%s", c.Year, c.Month))
	cursor, err := coll.
		Aggregate(context.Background(), pipeline)

	if err != nil {
		return 0, nil, err
	}
	defer cursor.Close(context.Background())

	var result []types.UserEvent
	for cursor.Next(context.Background()) {
		var log types.UserEvent
		_ = cursor.Decode(&log)
		result = append(result, log)
	}
	totalCount, err := coll.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, nil, err
	}
	return totalCount, result, nil
}
