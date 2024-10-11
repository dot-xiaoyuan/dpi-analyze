package member

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// IP 列表操作

type Tables struct {
	Results    []InfoJson `json:"results"`
	TotalCount int64      `json:"totalCount"`
}

type InfoJson struct {
	IP       string `json:"ip"`
	Mac      string `json:"mac"`
	TTL      string `json:"ttl"`
	UA       string `json:"ua"`
	LastSeen string `json:"last_seen"`
}

// TraversalIP 遍历IP表
func TraversalIP(startTime, endTime int64, page, pageSize int64) (result Tables, err error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (page - 1) * pageSize

	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	result.TotalCount = rdb.ZCount(ctx, types.ZSetIP, strconv.FormatInt(startTime, 10), strconv.FormatInt(endTime, 10)).Val()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, types.ZSetIP, &v9.ZRangeBy{
		Min:    strconv.FormatInt(startTime, 10), // 查询范围的最小时间戳
		Max:    strconv.FormatInt(endTime, 10),   // 查询范围的最大时间戳
		Offset: start,                            // 分页起始位置
		Count:  pageSize,                         // 每页大小
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("ZRange pipe.Exec", zap.Error(err))
		return
	}

	ips := zRangCmd.Val()
	// step2. 批量获取每个IP详细信息
	pipe = rdb.Pipeline()
	ipCommands := make([]*v9.MapStringStringCmd, 0, len(ips))
	for _, ip := range ips {
		key := fmt.Sprintf(types.HashAnalyzeIP, ip.Member.(string))
		cmd := pipe.HGetAll(ctx, key)
		ipCommands = append(ipCommands, cmd)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("HGetAll pipe.Exec", zap.Error(err))
		return
	}

	// step3. 处理查询结果
	ipDetails := make([]InfoJson, 0, len(ips))
	for i, cmd := range ipCommands {
		info := cmd.Val()
		detail := InfoJson{
			IP:       ips[i].Member.(string),
			TTL:      info["ttl"],
			UA:       info["ua"],
			Mac:      info["mac"],
			LastSeen: time.Unix(int64(ips[i].Score), 0).Format("2006/01/02 15:04:05"),
		}
		ipDetails = append(ipDetails, detail)
	}

	result.Results = ipDetails
	return
}
