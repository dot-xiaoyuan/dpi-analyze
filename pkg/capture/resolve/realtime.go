package resolve

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// 实时共享终端判定

func NewRealtime(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := types.ZSetRealtimeShored

	rdb.ZAdd(ctx, key, v9.Z{
		Score:  float64(time.Now().Unix()),
		Member: ip,
	}).Val()
}

func DeleteRealtime(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := types.ZSetRealtimeShored
	rdb.ZRem(ctx, key, ip).Val()
}

type Tables struct {
	Results    []InfoJson `json:"results"`
	TotalCount int64      `json:"totalCount"`
}

type InfoJson struct {
	IP       string  `json:"ip" redis:"ip"`
	Username string  `json:"username,omitempty" redis:"username,omitempty"`
	LastSeen float64 `json:"last_seen" redis:"last_seen"`
}

// TraversalIP 遍历IP表
func TraversalIP(startTime, endTime int64, page, pageSize int64) (result Tables, err error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (page - 1) * pageSize

	var setName string
	// 查询集合总数
	result.TotalCount = rdb.ZCount(ctx, types.ZSetRealtimeShored, strconv.FormatInt(startTime, 10), strconv.FormatInt(endTime, 10)).Val()
	if result.TotalCount == 0 {
		// 没数据直接返回
		result.Results = nil
		return
	}
	// Pipeline 批量查询
	pipe := rdb.Pipeline()
	// step1. 分页查询集合
	zRangCmd := rdb.ZRevRangeByScoreWithScores(ctx, setName, &v9.ZRangeBy{
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
	if len(ips) == 0 {
		result.Results = nil
		result.TotalCount = 0
		return
	}

	// step2. 批量获取每个IP详细信息
	pipe = rdb.Pipeline()
	ipCommands := make([]*v9.SliceCmd, 0, len(ips))

	for _, ip := range ips {
		key := fmt.Sprintf(types.HashAnalyzeIP, ip.Member.(string))
		ipCommands = append(ipCommands, pipe.HMGet(ctx, key, "username", "device"))
	}

	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, v9.Nil) {
		zap.L().Error("管道获取ip列表失败 pipe.Exec", zap.Error(err))
		return
	}

	// step3. 处理查询结果
	ipDetails := make([]InfoJson, 0, len(ips))
	for i, cmd := range ipCommands {
		//info := cmd.Val()
		var info InfoJson
		err = cmd.Scan(&info)
		if err != nil {
			zap.L().Error("Scan", zap.Error(err))
			continue
		}

		info.IP = ips[i].Member.(string)
		info.LastSeen = ips[i].Score
		ipDetails = append(ipDetails, info)
	}

	result.Results = ipDetails
	return
}
