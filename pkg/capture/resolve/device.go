package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
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
func storeDeviceInSet(ip string, device Device) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()

	// 将设备信息序列化
	deviceData := serializeDevice(device)

	// 将设备信息添加到该 IP 对应的设备集合中
	rdb.SAdd(ctx, ip, deviceData)

	// 设置过期时间（如 24 小时后过期）
	rdb.Expire(ctx, ip, 24*time.Hour)

	// 检查设备数量是否超过 1，触发事件
	checkAndTriggerEvent(ip)
}

// 触发事件函数
func triggerEvent(ip string) {
	// 这里可以是记录日志、发送通知或其他自定义处理逻辑
	log.Printf("Event Triggered: Multiple devices detected for IP %s", ip)
}

// 检查设备数量，并在满足条件时触发事件
func checkAndTriggerEvent(ip string) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()

	// 获取该 IP 下的设备数量
	deviceCount, err := rdb.SCard(ctx, ip).Result()
	if err != nil {
		log.Printf("Error getting device count for IP %s: %v", ip, err)
		return
	}

	// 如果设备数量超过 1，则触发事件
	if deviceCount > 1 {
		triggerEvent(ip)
	}
}

// 处理 SNI 解析结果（只存储厂商信息）
func handleSNI(ip string, brand string) {
	// 假设操作系统和型号信息未知
	device := Device{
		Manufacturer: strings.ToLower(brand),
		OS:           "unknown",
		Model:        "unknown",
	}

	// 存储设备信息到 Redis 集合中
	storeDeviceInSet(ip, device)
}

// 处理 UA 解析结果（更新操作系统和设备型号）
func handleUA(ip string, brand string, os string, model string) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()

	// 获取该 IP 下所有设备信息
	deviceData, err := rdb.SMembers(ctx, ip).Result()
	if err != nil {
		log.Printf("Error getting device data from Redis: %v", err)
		return
	}

	// 查看是否已存在该品牌的信息
	updated := false
	for _, data := range deviceData {
		var device Device
		err := json.Unmarshal([]byte(data), &device)
		if err != nil {
			log.Printf("Error deserializing device data: %v", err)
			continue
		}

		// 如果该品牌的信息已存在且操作系统为 unknown，则更新
		if device.Manufacturer == strings.ToLower(brand) && device.OS == "unknown" {
			// 更新操作系统和型号
			device.OS = os
			device.Model = model

			// 从集合中删除旧的设备信息
			rdb.SRem(ctx, ip, data)

			// 存储更新后的设备信息
			storeDeviceInSet(ip, device)

			updated = true
			break
		}
	}

	// 如果该 IP 下没有该品牌的信息，直接存储新的设备信息
	if !updated {
		newDevice := Device{
			Manufacturer: brand,
			OS:           os,
			Model:        model,
		}
		storeDeviceInSet(ip, newDevice)
	}
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
