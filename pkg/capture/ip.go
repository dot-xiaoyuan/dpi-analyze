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
	Results    []IPDetail `json:"results"`
	TotalCount int64      `json:"totalCount"`
}

type IPDetail struct {
	IP       string `json:"ip"`
	Mac      string `json:"mac"`
	TTL      string `json:"ttl"`
	UA       string `json:"ua"`
	LastSeen string `json:"last_seen"`
}

type TTLHistory struct {
	TTL       uint8     `bson:"ttl"`
	Timestamp time.Time `bson:"timestamp"`
}

type UAHistory struct {
	UserAgent string    `bson:"user_agent"`
	Timestamp time.Time `bson:"timestamp"`
}

type MacHistory struct {
	MacAddress string    `bson:"mac_address"`
	Timestamp  time.Time `bson:"timestamp"`
}

type IPActivityLogs struct {
	IP               string       `bson:"ip"`
	CurrentTTL       uint8        `bson:"current_ttl"`
	TTLHistory       []TTLHistory `bson:"ttl_history"`
	CurrentUserAgent string       `bson:"current_user_agent"`
	UAHistory        []UAHistory  `bson:"ua_history"`
	CurrentMac       string       `bson:"current_mac"`
	MacHistory       []MacHistory `bson:"mac_history"`
	LastSeen         time.Time    `bson:"last_seen"`
}

func StoreIPInZSet(ip string, timeStamp int64) {
	rdb := redis.GetRedisClient()
	rdb.ZAdd(context.TODO(), ZSetIPTable, redis2.Z{
		Score:  float64(timeStamp),
		Member: ip,
	})
}

func StoreIPInfoHash(ip, field string, value any) {
	if value == nil {
		return
	}
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(HashAnalyzeIP, ip)

	if field == "ttl" {
		newTTL := value.(uint8)
		oldTTL := rdb.HMGet(ctx, key, field).Val()
		if len(oldTTL) > 1 {
			for _, t := range oldTTL {
				if t != newTTL {
					value = append(oldTTL, newTTL)
					break
				}
			}
		}
	}
	rdb.HMSet(ctx, key, field, value, "last_seen", time.Now().Unix())
	rdb.Expire(ctx, key, 24*time.Hour)
}

func TraversalIP(start, stop int64) IPTables {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	result := IPTables{
		Results:    []IPDetail{},
		TotalCount: 0,
	}

	minScore := strconv.FormatInt(time.Now().Add(-24*time.Hour).Unix(), 10)
	maxScore := strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)
	result.TotalCount = rdb.ZCount(ctx, ZSetIPTable, minScore, maxScore).Val()

	zap.L().Info("condition", zap.String("minScore", minScore), zap.String("maxScore", maxScore), zap.Int64("start", start), zap.Int64("stop", stop))
	ips := rdb.ZRevRangeWithScores(ctx, ZSetIPTable, start, stop).Val()

	zap.L().Info("ips", zap.Any("ips", ips))
	for _, i := range ips {
		detail := getIPDetail(i.Member.(string))
		result.Results = append(result.Results, detail)
	}

	return result
}

func getIPDetail(ip string) IPDetail {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	detail := rdb.HGetAll(ctx, fmt.Sprintf(HashAnalyzeIP, ip)).Val()
	t, _ := strconv.ParseInt(detail["last_seen"], 10, 64)
	return IPDetail{
		IP:       ip,
		Mac:      detail["mac"],
		TTL:      detail["ttl"],
		UA:       detail["ua"],
		LastSeen: time.Unix(t, 0).Format("2006/01/02 15:04:05"),
	}
}

func persistToMongo() {

}
