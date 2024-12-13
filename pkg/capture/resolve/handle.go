package resolve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/parser"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strings"
	"time"
)

type DeviceInterface interface {
	serialize() string                              // json序列化
	unSerialize(jsonData string) types.DeviceRecord // json反序列化
	storeMongo()                                    // 录入mongo
	storeRedis(update bool)                         // 录入redis
	updateHash()                                    // 更新hash
	deleteRedis()                                   // 删除redis中设备集合
	checkDevice()                                   // 检查设备
	checkCount()                                    // 检查数量
}

type Device struct {
	IP     string
	Record types.DeviceRecord
	rdb    *v9.Client
	ctx    context.Context
}

// 将设备记录序列成json
func (d *Device) serialize() string {
	data, err := json.Marshal(d.Record)
	if err != nil {
		zap.L().Error("serialize device record", zap.Error(err))
		return ""
	}
	return string(data)
}

// 将json数据反序列成types.DeviceRecord
func (d *Device) unSerialize(jsonData string) types.DeviceRecord {
	var record types.DeviceRecord
	err := json.Unmarshal([]byte(jsonData), &record)
	if err != nil {
		zap.L().Error("un serialize device record", zap.Error(err))
		return types.DeviceRecord{}
	}
	return record
}

// 保存设备记录到mongodb
func (d *Device) storeMongo() {
	collection, err := createDynamicCollectionWithUniqueIndex(mongo.GetMongoClient(), types.MongoDatabaseDevices)
	if err != nil {
		zap.L().Error("Failed to create dynamic collection by devices", zap.Error(err))
		return
	}
	_, _ = collection.InsertOne(d.ctx, d.Record)
}

// 保存设备信息到redis
func (d *Device) storeRedis(update bool) {
	key := fmt.Sprintf(types.SetIPDevices, d.IP)

	if !update {
		storeDeviceIncr(d.rdb, d.Record)
	}
	// 序列化
	jsonData := d.serialize()
	if len(jsonData) == 0 {
		return
	}
	// 追加到集合中
	d.rdb.SAdd(d.ctx, key, jsonData).Val()
	// 设置过期时间
	d.rdb.Expire(d.ctx, key, 24*time.Hour).Val()
	// 追加设备信息到hash
	d.updateHash()
}

// 更新hash中的设备信息
func (d *Device) updateHash() {
	var brand string
	if len(d.Record.Brand) > 0 && d.Record.Brand != "Generic_Android" && d.Record.Brand != "android" {
		brand = strings.ToLower(d.Record.Brand)
	} else {
		brand = strings.ToLower(d.Record.Os)
	}
	zap.L().Debug("Wait to update hash", zap.String("brand", brand), zap.String("ip", d.IP))
	//
	var devices []parser.Domain
	domain := parser.Domain{
		Icon:      d.Record.Icon,
		BrandName: brand,
	}
	key := fmt.Sprintf(types.HashAnalyzeIP, d.IP)
	old := d.rdb.HMGet(d.ctx, key, string(types.Device)).Val()[0]
	if old != nil {
		// 如果存在旧数据，反序列化为设备列表
		if err := json.Unmarshal([]byte(old.(string)), &devices); err != nil {
			zap.L().Error("Failed to unmarshal devices from Redis", zap.String("key", key), zap.Error(err))
			return
		}

		// 遍历现有设备列表，检查是否有同品牌的记录
		updated := false
		for i, device := range devices {
			if device.BrandName == brand {
				// 如果品牌一致，更新设备信息
				if device.Icon != domain.Icon {
					devices[i].Icon = domain.Icon
				}
				updated = true
				break
			}
		}

		// 如果没有更新任何设备，说明是新设备，需要追加
		if !updated {
			devices = append(devices, domain)
		}
	} else {
		// 如果没有旧数据，初始化设备列表
		devices = append(devices, domain)
	}

	// 将更新后的设备列表序列化并写回 Redis
	bytes, err := json.Marshal(devices)
	if err != nil {
		zap.L().Error("Failed to marshal devices", zap.String("key", key), zap.Error(err))
		return
	}

	_, err = d.rdb.HSet(d.ctx, key, string(types.Device), bytes).Result()
	if err != nil {
		zap.L().Error("Failed to update Redis hash", zap.String("key", key), zap.Error(err))
	}
}

// 删除设备数量统计
func (d *Device) deleteRedis() {
	ctx := d.ctx
	d.rdb.Del(ctx, fmt.Sprintf(types.KeyDevicesAllIP, d.IP)).Val()
	d.rdb.Del(ctx, fmt.Sprintf(types.KeyDevicesMobileIP, d.IP)).Val()
	d.rdb.Del(ctx, fmt.Sprintf(types.KeyDevicesPcIP, d.IP)).Val()
}

// 检查设备信息
func (d *Device) checkDevice() {
	update := false
	key := fmt.Sprintf(types.SetIPDevices, d.IP)
	jsonData, err := d.rdb.SMembers(d.ctx, key).Result()
	if err != nil {
		zap.L().Error("check device failed", zap.Error(err))
		return
	}
	for _, device := range jsonData {
		oldRecord := d.unSerialize(device)
		if oldRecord.IP != d.IP {
			continue
		}
		zap.L().Info("diff", zap.Any("old", oldRecord), zap.Any("new", d.unSerialize(device)))
		// 操作系统一致，且版本不存在跳过
		if len(d.Record.Os) > 0 && d.Record.Os == oldRecord.Os && d.Record.Version == oldRecord.Version {
			update = true
			break
		}
		// 操作系统一致，但是品牌宽泛地跳过
		if d.Record.Os == oldRecord.Os &&
			d.Record.Version == oldRecord.Version &&
			(d.Record.Device == "Other" || d.Record.Brand == "android" || d.Record.Brand == "generic_android") {
			update = true
			break
		}
		// 联想
		if d.Record.Brand == "lenovo" && oldRecord.Os == "windows" && oldRecord.Brand == "windows" {
			// 更新操作系统和版本
			oldRecord.Brand = d.Record.Brand
			oldRecord.Icon = d.Record.Icon
			oldRecord.OriginChanel = d.Record.OriginChanel
			oldRecord.OriginValue = d.Record.OriginValue
			oldRecord.LastSeen = d.Record.LastSeen

			if len(d.Record.Device) > 0 && len(oldRecord.Device) == 0 {
				oldRecord.Device = d.Record.Device
			}
			if len(d.Record.Model) > 0 && len(oldRecord.Model) == 0 {
				oldRecord.Model = d.Record.Model
			}
			// 删除旧的设备信息
			d.rdb.SRem(d.ctx, key, device).Val()
			// 更新设备信息
			d.storeRedis(update)
			d.Record.Remark = "updated device"
			d.storeMongo()
			update = true
			break
		}
		// 品牌处理
		if len(d.Record.Brand) > 0 && oldRecord.Brand == d.Record.Brand {
			// 因为这里是更新品牌具体信息，如果操作系统和版本为空那么也跳过，仅当ua分析出来具体的系统和版本再进行设备的更新
			if len(d.Record.Os) == 0 || len(d.Record.Version) == 0 {
				update = true
				break
			}
			// 版本一致代表重复数据，也跳过
			if d.Record.Version == oldRecord.Version {
				update = true
				break
			}
			// TODO 其他条件 start

			// TODO end
			// 更新操作系统和版本
			oldRecord.Os = d.Record.Os
			oldRecord.Version = d.Record.Version
			oldRecord.OriginChanel = d.Record.OriginChanel
			oldRecord.OriginValue = d.Record.OriginValue
			oldRecord.LastSeen = d.Record.LastSeen

			if len(d.Record.Device) > 0 && len(oldRecord.Device) == 0 {
				oldRecord.Device = d.Record.Device
			}
			if len(d.Record.Model) > 0 && len(oldRecord.Model) == 0 {
				oldRecord.Model = d.Record.Model
			}
			// 删除旧的设备信息
			d.rdb.SRem(d.ctx, key, device).Val()
			// 更新设备信息
			d.storeRedis(update)
			d.Record.Remark = "updated device"
			d.storeMongo()
			update = true
			break
		}
	}

	if !update {
		d.storeRedis(update)
		d.Record.Remark = "saved device"
		d.storeMongo()
	}
}

// 检查设备数量
func (d *Device) checkCount() {
	key := fmt.Sprintf(types.SetIPDevices, d.IP)

	// 获取该 IP 下的设备数量
	deviceCount, err := d.rdb.SCard(d.ctx, key).Result()
	if err != nil {
		zap.L().Error("Error getting device count for IP %s: %v", zap.String("key", key), zap.Error(err))
		return
	}

	// 如果设备数量超过 1，则触发事件
	if deviceCount > 1 {
		triggerEvent(d.IP)
	}
}

var (
	expiration = 2 * time.Second
	retryCount = 5
	retryDelay = 100 * time.Millisecond
)

// Handle 设备处理
// 渠道 sni匹配 useragent匹配 ttl匹配
func Handle(device types.DeviceRecord) {
	rdb := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 增加上下文超时时间
	defer cancel()

	lockKey := fmt.Sprintf(types.LockIPBrand, device.IP, device.OriginChanel)

	// 获取锁
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	locked, err := acquireLock(ctx, rdb, lockKey, lockValue, expiration)
	if err != nil {
		zap.L().Error("Error acquiring lock", zap.String("key", lockKey), zap.Error(err), zap.String("IP", device.IP))
		return
	}
	if !locked {
		// zap.L().Warn("Failed to acquire lock", zap.String("key", lockKey), zap.String("IP", device.IP))
		return
	}

	// 确保锁在任务完成或上下文超时时释放
	defer func() {
		if err = releaseLock(context.Background(), rdb, lockKey, lockValue); err != nil {
			zap.L().Error("Error releasing lock", zap.String("key", lockKey), zap.Error(err), zap.String("IP", device.IP))
		}
	}()

	// 启动锁续约机制
	stopRenewal := startLockRenewal(ctx, rdb, lockKey, lockValue, expiration)
	defer stopRenewal()

	d := Device{
		IP:     device.IP,
		Record: device,
		rdb:    rdb,
		ctx:    ctx,
	}

	// 执行设备检查逻辑
	d.checkDevice()
}

func acquireLock(ctx context.Context, rdb *v9.Client, key, value string, expiration time.Duration) (bool, error) {
	locked, err := trySetLock(ctx, rdb, key, value, expiration)
	if err != nil {
		return false, err
	}
	if !locked {
		for i := 0; i < retryCount; i++ {
			time.Sleep(retryDelay)
			locked, err = trySetLock(ctx, rdb, key, value, expiration)
			if err != nil {
				return false, err
			}
			if locked {
				break
			}
		}
	}
	return locked, nil
}

func startLockRenewal(ctx context.Context, rdb *v9.Client, key, value string, expiration time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(expiration / 2) // 每半个过期时间续约一次
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := renewLock(ctx, rdb, key, value, expiration)
				if err != nil {
					zap.L().Error("Failed to renew lock", zap.String("key", key), zap.Error(err))
					return
				}
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

func renewLock(ctx context.Context, rdb *v9.Client, key, value string, expiration time.Duration) error {
	// 使用 Watch 监听 key，确保操作原子性
	err := rdb.Watch(ctx, func(tx *v9.Tx) error {
		// 获取当前锁的值
		currentValue, err := tx.Get(ctx, key).Result()
		if err != nil {
			if errors.Is(err, v9.Nil) {
				return fmt.Errorf("lock does not exist or expired")
			}
			return err
		}

		// 判断锁的所有权
		if currentValue != value {
			return fmt.Errorf("lock ownership mismatch")
		}

		// 续约
		_, err = tx.TxPipelined(ctx, func(pipe v9.Pipeliner) error {
			pipe.Expire(ctx, key, expiration)
			return nil
		})
		return err
	}, key)

	return err
}

// 尝试设置锁
func trySetLock(ctx context.Context, rdb *v9.Client, key, value string, expiration time.Duration) (bool, error) {
	return rdb.SetNX(ctx, key, value, expiration).Result()
}

// 释放锁
func releaseLock(ctx context.Context, rdb *v9.Client, key, value string) error {
	// 使用 Watch 和事务确保安全性
	err := rdb.Watch(ctx, func(tx *v9.Tx) error {
		currentValue, err := tx.Get(ctx, key).Result()
		if err != nil {
			if errors.Is(err, v9.Nil) {
				return nil // 锁已经不存在
			}
			return err
		}

		// 确认值匹配
		if currentValue != value {
			return nil // 不释放其他客户端的锁
		}

		// 删除锁
		_, err = tx.TxPipelined(ctx, func(pipe v9.Pipeliner) error {
			pipe.Del(ctx, key)
			return nil
		})
		return err
	}, key)

	return err
}
