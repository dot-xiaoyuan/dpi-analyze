package resolve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/parser"
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

// 将设备信息序列化为 JSON 字符串
func serializeDevice(device types.DeviceRecord) string {
	data, err := json.Marshal(device)
	if err != nil {
		log.Printf("Error serializing device: %v", err)
		return ""
	}
	return string(data)
}

// 保存设备信息
func storeDevice(rdb *v9.Client, device types.DeviceRecord, update bool) {
	key := fmt.Sprintf(types.SetIPDevices, device.IP)

	if !update {
		storeDeviceIncr(rdb, device)
	}
	// 将设备信息序列化
	deviceData := serializeDevice(device)

	// 将设备信息添加到该 IP 对应的设备集合中
	rdb.SAdd(context.TODO(), key, deviceData).Val()

	// 设置过期时间（如 24 小时后过期）
	rdb.Expire(context.TODO(), key, 24*time.Hour).Val()

	// 设置hash
	var str string
	if len(device.Brand) > 0 && device.Brand != "Generic_Android" && device.Brand != "android" {
		str = strings.ToLower(device.Brand)
	} else {
		str = strings.ToLower(device.Os)
	}
	zap.L().Debug("wait to save device", zap.String("key", key), zap.String("str", str))
	AppendDevice2Redis(device.IP, types.Device, str, device)

	// 检查设备数量是否超过 1，触发事件
	checkAndTriggerEvent(device.IP)
}

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
	zap.L().Warn("Event Triggered: Multiple devices detected for IP ", zap.String("ip", ip))
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

// DeviceHandle 设备handle
func DeviceHandle(device types.DeviceRecord) {
	rdb := redis.GetRedisClient()
	ctx := context.Background()
	key := fmt.Sprintf(types.SetIPDevices, device.IP)
	lockKey := fmt.Sprintf(types.LockIPBrand, device.IP, device.Brand)

	// 尝试获取锁
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	locked, err := rdb.SetNX(ctx, lockKey, lockValue, time.Second).Result()
	if err != nil {
		zap.L().Error("Error acquiring Redis lock: %v", zap.Error(err))
		return
	}
	if !locked {
		retryCount := 5
		retryDelay := 100 * time.Millisecond
		for i := 0; i < retryCount; i++ {
			time.Sleep(retryDelay)
			locked, err = rdb.SetNX(ctx, lockKey, lockValue, time.Second).Result()
			if err != nil {
				zap.L().Error("Error re-trying to acquire Redis lock: %v", zap.Error(err))
				return
			}
			if locked {
				break
			}
		}

		if !locked {
			// 如果仍然没有获取到锁，记录一次日志
			// zap.L().Warn("Another process is handling the same brand and IP, skipping.", zap.String("ip", device.IP))
			return
		}

	}
	defer func() {
		val, _ := rdb.Get(ctx, lockKey).Result()
		if val == lockValue {
			err = rdb.Del(ctx, lockKey).Err()
			if err != nil {
				zap.L().Error("Error releasing Redis lock: %v", zap.Error(err))
			}
		}
	}()

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
		// 如果系统一样，比对版本
		if len(device.Os) > 0 && d.Os == device.Os && len(device.Version) == 0 {
			updated = true
			break
		}
		// 如果该品牌的信息已存在且操作系统为 unknown，则更新
		if len(device.Brand) > 0 && d.Brand == device.Brand {
			zap.L().Debug("device", zap.Any("device", device), zap.Any("d", d))
			if device.Os == "" || device.Version == "" {
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

			if device.Device != "" && d.Device == "" {
				d.Device = device.Device
			}
			if device.Model != "" && d.Model == "" {
				d.Model = device.Model
			}

			// 从集合中删除旧的设备信息
			rdb.SRem(ctx, key, data)

			// 存储更新后的设备信息
			storeDevice(rdb, d, true)
			d.Remark = "updated device"
			storeMongo(d)

			updated = true
			break
		}
		// 忽略系统版本一致但是设备是 Other 或者 品牌是 android的
		if d.Os == device.Os && d.Version == device.Version && (device.Device == "Other" || device.Brand == "android" || device.Brand == "generic_android") {
			updated = true
			break
		}
	}

	// 如果该 IP 下没有该品牌的信息，直接存储新的设备信息
	if !updated {
		storeDevice(rdb, device, false)
		device.Remark = "saved device"
		storeMongo(device)
	}
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

// AppendDevice2Redis 追加设备信息到redis
func AppendDevice2Redis(ip string, property types.Property, value any, dr types.DeviceRecord) {
	rdb := redis.GetRedisClient()
	ctx := context.TODO()
	key := fmt.Sprintf(types.HashAnalyzeIP, ip)

	var devices []parser.Domain

	_, mf := features.HandleFeatureMatch(value.(string), ip, dr)
	// info hash
	old := rdb.HMGet(ctx, key, string(property)).Val()[0]
	if old != nil {
		_ = json.Unmarshal([]byte(old.(string)), &devices)
		for i, device := range devices {
			if device.BrandName == value {
				if device.Icon != mf.Icon {
					devices = append(devices[:i], devices[i+1:]...)
				}
				return
			}
		}
		devices = append(devices, mf)
		bytes, _ := json.Marshal(devices)
		rdb.HSet(ctx, key, string(property), bytes).Val()
	} else {
		devices = append(devices, mf)
		bytes, _ := json.Marshal(devices)
		rdb.HSet(ctx, key, string(property), bytes).Val()
	}
}
