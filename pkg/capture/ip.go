package capture

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	redis2 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type IPTables struct {
	Results    []IPInfoOld `json:"results"`
	TotalCount int64       `json:"totalCount"`
}

type IPInfoOld struct {
	IP       string `json:"ip"`
	Mac      string `json:"mac"`
	TTL      string `json:"ttl"`
	UA       string `json:"ua"`
	LastSeen string `json:"last_seen"`
}

//	type TTLHistory struct {
//		TTL       uint8     `bson:"ttl"`
//		Timestamp time.Time `bson:"timestamp"`
//	}
//
//	type UAHistory struct {
//		UserAgent string    `bson:"user_agent"`
//		Timestamp time.Time `bson:"timestamp"`
//	}
//
//	type MacHistory struct {
//		MacAddress string    `bson:"mac_address"`
//		Timestamp  time.Time `bson:"timestamp"`
//	}
//
//	type IPActivityLogs struct {
//		IP               string       `bson:"ip"`
//		CurrentTTL       uint8        `bson:"current_ttl"`
//		TTLHistory       []TTLHistory `bson:"ttl_history"`
//		CurrentUserAgent string       `bson:"current_user_agent"`
//		UAHistory        []UAHistory  `bson:"ua_history"`
//		CurrentMac       string       `bson:"current_mac"`
//		MacHistory       []MacHistory `bson:"mac_history"`
//		LastSeen         time.Time    `bson:"last_seen"`
//	}
//
// // StoreIPInZSet 加载IP至有序集合
//
//	func StoreIPInZSet(ip string, timeStamp int64) {
//		rdb := redis.GetRedisClient()
//		rdb.ZAdd(context.TODO(), ZSetIPTable, redis2.Z{
//			Score:  float64(timeStamp),
//			Member: ip,
//		})
//	}
//
// // StoreIPInfoHash 加载IP详情hash
func StoreIPInfoHash(ip, field string, value any) {
	if value == nil {
		return
	}
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(HashAnalyzeIP, ip)

	var newVal any
	if field == "ttl" {
		newVal = value.(uint8)
	} else {
		newVal = value.(string)
	}
	oldTTL := rdb.HMGet(ctx, key, field).Val()
	if len(oldTTL) > 1 {
		for _, t := range oldTTL {
			if t != newVal {
				value = append(oldTTL, newVal)
				break
			}
		}
	}
	rdb.HMSet(ctx, key, field, value, "last_seen", time.Now().Unix())
	rdb.Expire(ctx, key, 24*time.Hour)
}

// TraversalIP 遍历IP表
func TraversalIP(startTime, endTime int64, page, pageSize int64) ([]IPInfoOld, error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (page - 1) * pageSize

	// Pipeline 批量查询
	pipe := rdb.Pipeline()

	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, ZSetIPTable, &redis2.ZRangeBy{
		Min:    strconv.FormatInt(startTime, 10), // 查询范围的最小时间戳
		Max:    strconv.FormatInt(endTime, 10),   // 查询范围的最大时间戳
		Offset: start,                            // 分页起始位置
		Count:  pageSize,                         // 每页大小
	})

	_, err := pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("ZRange pipe.Exec", zap.Error(err))
		return nil, err
	}

	ips := zRangCmd.Val()
	// step2. 批量获取每个IP详细信息
	pipe = rdb.Pipeline()
	ipCommands := make([]*redis2.MapStringStringCmd, 0, len(ips))
	for _, ip := range ips {
		key := fmt.Sprintf(HashAnalyzeIP, ip.Member.(string))
		cmd := pipe.HGetAll(ctx, key)
		ipCommands = append(ipCommands, cmd)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("HGetAll pipe.Exec", zap.Error(err))
		return nil, err
	}

	// step3. 处理查询结果
	ipDetails := make([]IPInfoOld, 0, len(ips))
	for i, cmd := range ipCommands {
		info := cmd.Val()
		detail := IPInfoOld{
			IP:       ips[i].Member.(string),
			TTL:      info["ttl"],
			UA:       info["ua"],
			Mac:      info["mac"],
			LastSeen: time.Unix(int64(ips[i].Score), 0).Format("2006/01/02 15:04:05"),
		}
		ipDetails = append(ipDetails, detail)
	}

	return ipDetails, nil
}

func persistToMongo() {

}
