package member

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/ants"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 特征

var (
	featureCaches sync.Map
	qpsCounters   sync.Map
	weights       sync.Map
)

type TimeFeature struct {
	Timestamp time.Time
	Value     string
}

type Feature struct {
	IP    string
	Field types.Feature
	Value string
}

type IPTimeWindow struct {
	IP   string
	Data []FeatureData
}
type FeatureData struct {
	Field    types.Feature
	Features []TimeFeature
}

func GetCache(field types.Feature) *sync.Map {
	cache, _ := featureCaches.LoadOrStore(field, &sync.Map{})
	return cache.(*sync.Map)
}

func Increment(feature Feature) {
	m := GetCache(feature.Field)
	now := time.Now()

	features, ok := GetFeatureByMemory(feature.IP, m)
	if ok {
		// 判断时间窗口,如果超过时间窗口 清空缓存
		// 如果特征值已存在并在时间窗口内，则跳过
		for _, f := range features {
			if f.Value == feature.Value && now.Sub(f.Timestamp) < (config.Cfg.Feature.SNI.TimeWindow*time.Second) {
				return
			}
		}
	}

	// 添加or更新特征
	newFeature := TimeFeature{Value: feature.Value, Timestamp: now}
	features = append(features, newFeature)
	putFeatureByMemory(feature.IP, m, features)

	zap.L().Debug("Increment", zap.Any("feature", feature))
}

func GetFeatureByMemory(ip string, m *sync.Map) ([]TimeFeature, bool) {
	val, ok := m.Load(ip)
	if ok {
		return val.([]TimeFeature), true
	}
	return nil, false
}

func putFeatureByMemory(ip string, m *sync.Map, values []TimeFeature) {
	m.Store(ip, values)
}

// IncrementQPS 增加 QPS 计数
func IncrementQPS(ip string) {
	count, _ := qpsCounters.LoadOrStore(ip, 0)
	qpsCounters.Store(ip, count.(int)+1)
}

func GetQPS(ip string) int {
	if qps, ok := qpsCounters.Load(ip); ok {
		return qps.(int)
	}
	return 0
}

// CleanExpiredData 清理过期数据
// 每次清理前将数据保存到 mongodb
func CleanExpiredData() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		featureCaches.Range(func(key, value any) bool {
			m := value.(*sync.Map)
			field := key.(types.Feature)

			// 收集并保存IP所有特征数据
			var itw IPTimeWindow
			// 记录时间窗口内缓存的数据，然后清空缓存
			m.Range(func(ip, features any) bool {
				if itw.IP == "" {
					itw.IP = ip.(string)
				}
				itw.Data = append(itw.Data, FeatureData{
					Field:    field,
					Features: features.([]TimeFeature),
				})
				// 清理缓存
				m.Delete(ip)
				return true
			})
			_ = ants.Submit(func() {
				save2mongo(itw)
			})
			return true
		})
		// 每分钟重置 QPS 计数器
		resetQPSCounters()
	}
}

// 重置 QPS 计数器
func resetQPSCounters() {
	qpsCounters.Range(func(key, _ any) bool {
		qpsCounters.Store(key, 0)
		return true
	})
}

// PrintStatistics 输出统计信息
func PrintStatistics() {
	weights.Range(func(ip, weight any) bool {
		qps, _ := qpsCounters.Load(ip)
		fmt.Printf("IP: %s, QPS: %d, Weight: %d\n", ip, qps.(int), weight.(int))
		return true
	})
}

func save2mongo(itw IPTimeWindow) {
	if !config.UseMongo {
		return
	}
	raw := make(map[string]interface{})
	// 统计数量 (一个时间窗口内的数量)
	for _, item := range itw.Data {
		raw[string(item.Field)] = item.Features
		raw[fmt.Sprintf("%s_count", item.Field)] = len(item.Features)
	}

	err := mongo.Mongo.InsertOneFeature(itw.IP, raw)
	if err != nil {
		panic(err)
	}
}
