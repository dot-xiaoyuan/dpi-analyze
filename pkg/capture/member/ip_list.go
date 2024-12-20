package member

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

// IP 列表操作

type Tables struct {
	Results    []InfoJson `json:"results"`
	TotalCount int64      `json:"totalCount"`
}

type InfoJson struct {
	IP         string  `json:"ip" redis:"ip"`
	Username   string  `json:"username,omitempty" redis:"username,omitempty"`
	Mac        string  `json:"mac,omitempty" redis:"mac,omitempty"`
	TTL        string  `json:"ttl,omitempty" redis:"ttl,omitempty"`
	UA         string  `json:"user_agent,omitempty" redis:"user_agent,omitempty"`
	Device     string  `json:"device,omitempty" redis:"device,omitempty"`
	DeviceName string  `json:"device_name" redis:"device_name,omitempty"`
	DeviceType string  `json:"device_type" redis:"device_type,omitempty"`
	All        string  `json:"all,omitempty" redis:"all,omitempty"`
	Mobile     string  `json:"mobile,omitempty" redis:"mobile,omitempty"`
	Pc         string  `json:"pc,omitempty" redis:"pc,omitempty"`
	LastSeen   float64 `json:"last_seen" redis:"last_seen"`
}

// TraversalIP 遍历IP表
func TraversalIP(startTime, endTime int64, page, pageSize int64) (result Tables, err error) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()

	// 分页的起止索引
	start := (page - 1) * pageSize

	var setName string
	if config.FollowOnlyOnlineUsers {
		setName = types.ZSetOnlineUsers
	} else {
		setName = types.ZSetIP
	}
	// 查询集合总数
	result.TotalCount = rdb.ZCount(ctx, setName, strconv.FormatInt(startTime, 10), strconv.FormatInt(endTime, 10)).Val()
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
	devicesCommands := make([]*v9.StringSliceCmd, 0, len(ips))
	extraKeyCommands := make([]*v9.StringCmd, 0, len(ips)*3)

	for _, ip := range ips {
		key := fmt.Sprintf(types.HashAnalyzeIP, ip.Member.(string))
		//cmd := pipe.HGetAll(ctx, key)
		ipCommands = append(ipCommands, pipe.HMGet(ctx, key, "username", "mac", "ttl", "user_agent", "device_name", "device_type"))

		// 获取设备集合
		devicesCommands = append(devicesCommands, pipe.SMembers(ctx, fmt.Sprintf(types.SetIPDevices, ip.Member.(string))))
		// 获取设备数量
		allKey := pipe.Get(ctx, fmt.Sprintf(types.KeyDevicesAllIP, ip.Member.(string)))
		mobileKey := pipe.Get(ctx, fmt.Sprintf(types.KeyDevicesMobileIP, ip.Member.(string)))
		pcKey := pipe.Get(ctx, fmt.Sprintf(types.KeyDevicesPcIP, ip.Member.(string)))
		extraKeyCommands = append(extraKeyCommands, allKey, mobileKey, pcKey)
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
		// 设备列表
		devices := devicesCommands[i].Val()

		all := extraKeyCommands[i*3].Val()
		mobile := extraKeyCommands[i*3+1].Val()
		pc := extraKeyCommands[i*3+2].Val()

		info.Device = "[" + strings.Join(devices, ",") + "]"
		info.All, info.Mobile, info.Pc = all, mobile, pc
		info.IP = ips[i].Member.(string)
		info.LastSeen = ips[i].Score
		ipDetails = append(ipDetails, info)
	}

	result.Results = ipDetails
	return
}
