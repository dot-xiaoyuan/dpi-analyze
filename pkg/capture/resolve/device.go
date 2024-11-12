package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

// Device 设备信息结构体
type Device struct {
	Manufacturer string `json:"brand"`
	OS           string `json:"os"`
	Model        string `json:"model"`
}

// 将设备信息序列化为 JSON 字符串
func serializeDevice(device Device) string {
	data, err := json.Marshal(device)
	if err != nil {
		log.Printf("Error serializing device: %v", err)
		return ""
	}
	return string(data)
}

// 存储设备信息到 Redis 集合中，并检查设备数量是否超过 1
func checkSet(ip string, device Device) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()
	key := fmt.Sprintf(types.SetIPDevices, ip)

	// 获取该 IP 下所有设备信息
	deviceData, err := rdb.SMembers(ctx, key).Result()
	if err != nil {
		zap.L().Error("Error getting device data from Redis: %v", zap.Error(err))
		return
	}

	// 查看是否已存在该品牌的信息
	updated := false
	for _, data := range deviceData {
		var d Device
		err = json.Unmarshal([]byte(data), &d)
		if err != nil {
			zap.L().Error("Error deserializing device data: %v", zap.Error(err))
			continue
		}

		// 如果该品牌的信息已存在且操作系统为 unknown，则更新
		if d.Manufacturer == strings.ToLower(device.Manufacturer) && d.OS == "unknown" {
			// 更新操作系统和型号
			d.OS = device.OS
			d.Model = device.Model

			// 从集合中删除旧的设备信息
			rdb.SRem(ctx, key, data)

			// 存储更新后的设备信息
			storeDevice(rdb, ip, d)

			updated = true
			break
		}
	}

	// 如果该 IP 下没有该品牌的信息，直接存储新的设备信息
	if !updated {
		newDevice := Device{
			Manufacturer: strings.ToLower(device.Manufacturer),
			OS:           device.OS,
			Model:        device.Model,
		}
		storeDevice(rdb, ip, newDevice)
	}
}

func storeDevice(rdb *v9.Client, ip string, device Device) {
	key := fmt.Sprintf(types.SetIPDevices, ip)

	// 将设备信息序列化
	deviceData := serializeDevice(device)

	// 将设备信息添加到该 IP 对应的设备集合中
	rdb.SAdd(context.TODO(), key, deviceData)

	// 设置过期时间（如 24 小时后过期）
	rdb.Expire(context.TODO(), key, 24*time.Hour)

	// 检查设备数量是否超过 1，触发事件
	checkAndTriggerEvent(ip)
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

// 处理 SNI 解析结果（只存储厂商信息）
func handleSNI(ip string, brand string) {
	checkSet(ip, Device{
		Manufacturer: strings.ToLower(brand),
		OS:           "unknown",
		Model:        "unknown",
	})
}

// 处理 UA 解析结果（更新操作系统和设备型号）
func handleUA(ip string, brand string, os string, model string) {
	checkSet(ip, Device{
		Manufacturer: strings.ToLower(brand),
		OS:           os,
		Model:        model,
	})
}

// 获取某个 IP 下的所有设备信息
func getDevicesByIP(ip string) ([]Device, error) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()

	// 获取该 IP 对应的所有设备信息
	deviceData, err := rdb.SMembers(ctx, ip).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices for IP %s: %v", ip, err)
	}

	var devices []Device
	for _, data := range deviceData {
		var device Device
		err := json.Unmarshal([]byte(data), &device)
		if err != nil {
			log.Printf("Error deserializing device data: %v", err)
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// ProcessRequest 处理请求，SNI 和 UA 解析结果的入口
func ProcessRequest(ip string, brand string, os string, model string) {
	// 如果只有厂商信息（SNI解析）
	if os == "" && model == "" {
		handleSNI(ip, brand)
	} else {
		// 如果有完整的 UA 信息
		handleUA(ip, brand, os, model)
	}
}
