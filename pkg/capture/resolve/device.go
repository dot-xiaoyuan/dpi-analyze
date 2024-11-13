package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/member"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

// 将设备信息序列化为 JSON 字符串
func serializeDevice(device types.DeviceRecord) string {
	data, err := json.Marshal(device)
	if err != nil {
		log.Printf("Error serializing device: %v", err)
		return ""
	}
	return string(data)
}

func storeDevice(rdb *v9.Client, device types.DeviceRecord) {
	key := fmt.Sprintf(types.SetIPDevices, device.IP)

	// 将设备信息序列化
	deviceData := serializeDevice(device)

	// 将设备信息添加到该 IP 对应的设备集合中
	rdb.SAdd(context.TODO(), key, deviceData)

	// 设置过期时间（如 24 小时后过期）
	rdb.Expire(context.TODO(), key, 24*time.Hour)

	// 设置hash
	var str string
	if len(device.Brand) == 0 || device.Brand == "unknown" {
		str = strings.ToLower(device.Os)
	} else {
		str = strings.ToLower(device.Brand)
	}
	member.AppendDevice2Redis(device.IP, types.Device, str)

	// 检查设备数量是否超过 1，触发事件
	checkAndTriggerEvent(device.IP)
}

func storeMongo(device types.DeviceRecord) {
	_, _ = mongo.GetMongoClient().Database(string(types.Device)).Collection("record").InsertOne(context.TODO(), device)
}

// 触发事件函数
func triggerEvent(ip string) {
	zap.L().Warn("Event Triggered: Multiple devices detected for IP ", zap.String("ip", ip))
}

// 检查设备数量，并在满足条件时触发事件
func checkAndTriggerEvent(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()
	key := fmt.Sprintf(types.SetIPDevices, ip)

	// 获取该 IP 下的设备数量
	deviceCount, err := rdb.SCard(ctx, key).Result()
	if err != nil {
		zap.L().Error("Error getting device count for IP %s: %v", zap.String("key", key), zap.Error(err))
		return
	}

	// 如果设备数量超过 1，则触发事件
	if deviceCount > 1 {
		triggerEvent(ip)
	}
}

// GetDevicesByIP 获取某个 IP 下的所有设备信息
func GetDevicesByIP(ip string) ([]types.DeviceRecord, error) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()

	// 获取该 IP 对应的所有设备信息
	deviceData, err := rdb.SMembers(ctx, fmt.Sprintf(types.SetIPDevices, ip)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices for IP %s: %v", ip, err)
	}

	var devices []types.DeviceRecord
	for _, data := range deviceData {
		var device types.DeviceRecord
		err = json.Unmarshal([]byte(data), &device)
		if err != nil {
			log.Printf("Error deserializing device data: %v", err)
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// DeviceHandle 设备handle
func DeviceHandle(device types.DeviceRecord) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()
	key := fmt.Sprintf(types.SetIPDevices, device.IP)

	// 获取该 IP 下所有设备信息
	deviceData, err := rdb.SMembers(ctx, key).Result()
	if err != nil {
		zap.L().Error("Error getting device data from Redis: %v", zap.Error(err))
		return
	}

	// 查看是否已存在该品牌的信息
	updated := false
	for _, data := range deviceData {
		var d types.DeviceRecord
		err = json.Unmarshal([]byte(data), &d)
		if err != nil {
			zap.L().Error("Error deserializing device data: %v", zap.Error(err))
			continue
		}
		// 如果该品牌的信息已存在且操作系统为 unknown，则更新
		if d.Brand == strings.ToLower(device.Brand) {
			if device.Os == "unknown" || device.Version == "unknown" {
				// 重复的unknown跳过
				updated = true
				break
			}
			if device.Version == d.Version {
				// 相同版本跳过
				updated = true
				break
			}
			// 更新操作系统和型号
			d.Os = device.Os
			d.Version = device.Version

			d.OriginChanel = device.OriginChanel
			d.OriginValue = device.OriginValue
			d.LastSeen = device.LastSeen

			if device.Device != "unknown" && d.Device == "unknown" {
				d.Device = device.Device
			}
			if device.Model != "unknown" && d.Model == "unknown" {
				d.Model = device.Model
			}

			// 从集合中删除旧的设备信息
			rdb.SRem(ctx, key, data)

			// 存储更新后的设备信息
			storeDevice(rdb, d)

			updated = true
			break
		}
	}

	// 如果该 IP 下没有该品牌的信息，直接存储新的设备信息
	if !updated {
		storeDevice(rdb, device)
		storeMongo(device)
	}
}
