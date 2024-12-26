package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	v9 "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	triggerLock sync.Mutex
)

// 创建索引并返回集合
func createDynamicCollectionWithUniqueIndex(client *driver.Client, dbName string) (*driver.Collection, error) {
	// 根据时间动态生成集合名
	collectionName := time.Now().Format("06_01_02")
	collection := client.Database(dbName).Collection(collectionName)

	// 检查复合唯一索引是否已存在
	if err := ensureUniqueCompoundIndexExists(collection); err != nil {
		return nil, fmt.Errorf("failed to check if index exists: %w", err)
	}

	return collection, nil
}

// 检查索引是否存在
func ensureUniqueCompoundIndexExists(collection *driver.Collection) error {
	// 检查是否存在名为 "unique_ip_last_seen_index" 的索引
	cursor, err := collection.Indexes().List(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return fmt.Errorf("failed to decode index: %w", err)
		}

		// 查找具有指定名称的索引
		if index["name"] == "unique_ip_last_seen_index" {
			// 如果索引已经存在，直接返回
			return nil
		}
	}

	// 如果没有找到指定名称的索引，创建复合唯一索引
	indexModel := driver.IndexModel{
		Keys: bson.D{
			{Key: "ip", Value: 1},        // 升序索引
			{Key: "last_seen", Value: 1}, // 升序索引
		},
		Options: options.Index().
			SetUnique(true).                      // 设置唯一索引
			SetName("unique_ip_last_seen_index"), // 给索引指定一个名称
	}

	_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
	return err
}

// 设备录入记录
func storeMongo(device types.DeviceRecord) {
	collection, err := createDynamicCollectionWithUniqueIndex(mongo.GetMongoClient(), types.MongoDatabaseDevices)
	if err != nil {
		zap.L().Error("failed to create unique index", zap.Error(err))
		return
	}
	_, _ = collection.InsertOne(context.TODO(), device)
}

// 触发事件函数
func triggerEvent(ip string) {
	//zap.L().Warn("Event Triggered: Multiple devices detected for IP ", zap.String("ip", ip))
	triggerLock.Lock()
	Discover(ip)
	triggerLock.Unlock()
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

// 保存设备数
func storeDeviceIncr(rdb *v9.Client, device types.DeviceRecord) {
	// ALL ++
	rdb.Incr(context.TODO(), fmt.Sprintf(types.KeyDevicesAllIP, device.IP)).Val()

	if IsMobile(device) {
		rdb.Incr(context.TODO(), fmt.Sprintf(types.KeyDevicesMobileIP, device.IP)).Val()
	} else {
		rdb.Incr(context.TODO(), fmt.Sprintf(types.KeyDevicesPcIP, device.IP))
	}
}

// GetDeviceIncr 获取设备数量
func GetDeviceIncr(ip string, rdb *v9.Client) (all, mobile, pc int) {
	key := []string{fmt.Sprintf(types.KeyDevicesAllIP, ip), fmt.Sprintf(types.KeyDevicesMobileIP, ip), fmt.Sprintf(types.KeyDevicesPcIP, ip)}

	values, err := rdb.MGet(context.TODO(), key...).Result()
	if err != nil {
		zap.L().Error("Error getting device incr", zap.String("ip", ip), zap.Error(err))
		return 0, 0, 0
	}
	if values[0] != nil {
		all, _ = strconv.Atoi(values[0].(string))
	}
	if values[1] != nil {
		mobile, _ = strconv.Atoi(values[1].(string))
	}
	if values[2] != nil {
		pc, _ = strconv.Atoi(values[2].(string))
	}
	return
}

// DelDeviceIncr 删除设备数量信息
func DelDeviceIncr(ip string, rdb *v9.Client) {
	rdb.Del(context.TODO(), fmt.Sprintf(types.KeyDevicesAllIP, ip)).Val()
	rdb.Del(context.TODO(), fmt.Sprintf(types.KeyDevicesMobileIP, ip)).Val()
	rdb.Del(context.TODO(), fmt.Sprintf(types.KeyDevicesPcIP, ip)).Val()
}

// IsMobile 判断客户端是否为移动设备
func IsMobile(device types.DeviceRecord) bool {
	if device.Device != "" {
		deviceFamily := strings.ToLower(device.Device)
		// 判断 Device 家族是否包含常见移动设备标识
		if strings.Contains(deviceFamily, "phone") || strings.Contains(deviceFamily, "mobile") {
			return true
		}
		if strings.Contains(deviceFamily, "windows") {
			return false
		}
	}

	if device.Os != "" {
		osFamily := strings.ToLower(device.Os)
		// 判断 OS 是否为常见的移动操作系统
		if strings.Contains(osFamily, "android") || strings.Contains(osFamily, "ios") {
			return true
		}
		if strings.Contains(osFamily, "mac") || strings.Contains(osFamily, "windows") {
			return false
		}
	}

	// UserAgent 通常作为辅助判断，可以补充更多规则
	if device.OriginValue != "" {
		uaFamily := strings.ToLower(device.OriginValue)
		// 判断 UserAgent 是否明显表明是移动设备
		if strings.Contains(uaFamily, "mobile") || strings.Contains(uaFamily, "phone") {
			return true
		}
	}

	// 默认返回 PC
	return true
}
