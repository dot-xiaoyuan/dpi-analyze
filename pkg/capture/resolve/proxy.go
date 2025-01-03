package resolve

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/policy"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// 代理handle

func NewRecord(ip, username string, devices []types.DeviceRecord) *types.ProxyRecord {
	pr := &types.ProxyRecord{
		IP:       ip,
		Username: username,
		Devices:  devices,
		LastSeen: time.Now(),
	}
	return pr
}

func HandleProxy(record *types.ProxyRecord) {
	_, _ = mongo.GetMongoClient().Database(types.MongoDatabaseProxy).
		Collection(time.Now().Format("06_01")).
		InsertOne(context.TODO(), record)
}

// Discover 检测到当前设备数异常后处理
func Discover(ip string) {
	// 时间间隔，如果短时间内处理过
	rdb := redis.GetRedisClient()
	ctx := context.Background()
	key := fmt.Sprintf(types.KeyDiscoverIP, ip)

	ttl := rdb.TTL(ctx, key).Val()
	if ttl > 0 {
		return
	}
	// 获取用户详情
	user := users.FindUser(ip)
	if user.UserName == "" {
		afterDiscover(key, rdb)
		zap.L().Warn("用户不存在", zap.String("ip", ip))
		return
	}
	// 记录到实时共享终端判定记录中
	NewRealtime(ip)
	// TODO 获取控制策略,检测是否开启防代理

	// 获取产品对应条件
	conditionAll, conditionMobile, conditionPc, disable := getStrategyByProduct(user.ProductsID)
	// 获取设备信息
	all, mobile, pc := GetDeviceIncr(ip, rdb)
	if all < conditionAll && mobile < conditionMobile && pc < conditionPc {
		zap.L().Warn("设备数量不满足判定条件", zap.String("ip", ip), zap.Int("mobile", mobile), zap.Int("pc", pc), zap.Int("all", all))
		return
	}
	// 满足代理条件
	devices, err := GetDevicesByIP(ip)
	if err != nil {
		zap.L().Error("获取用户设备信息失败")
		afterDiscover(key, rdb)
		return
	}
	pr := NewRecord(ip, user.UserName, devices)
	pr.AllCount, pr.MobileCount, pr.PcCount = all, mobile, pc
	if disable == 1 {
		_ = users.HookDropUser(user, pr)
	}
	HandleProxy(pr)
	DelDeviceIncr(ip, rdb)
	afterDiscover(key, rdb)
}

// 根据产品获取对应的策略
func getStrategyByProduct(productID int) (all, mobile, pc, disable int) {
	product := policy.Get(strconv.Itoa(productID))
	all, mobile, pc, disable = product.ALL, product.Mobile, product.Pc, product.DisableProxy
	return
}

func afterDiscover(key string, rdb *v9.Client) {
	rdb.Set(context.Background(), key, time.Now(), time.Minute*5).Val()
}
