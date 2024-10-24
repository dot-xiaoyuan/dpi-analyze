package member

import (
	"fmt"
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
func CleanExpiredData() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now()
		featureCaches.Range(func(key, value any) bool {
			m := value.(*sync.Map)

			// 清理每个特征缓存中的过期数据
			m.Range(func(ip, features any) bool {
				filtered := filterExpired(features.([]TimeFeature), now)
				if len(filtered) > 0 {
					m.Store(ip, features)
				} else {
					m.Delete(ip)
				}
				return true
			})
			return true
		})
		// 每分钟重置 QPS 计数器
		resetQPSCounters()
	}
}

// 过滤过期数据
func filterExpired(features []TimeFeature, now time.Time) []TimeFeature {
	var result []TimeFeature
	for _, f := range features {
		if now.Sub(f.Timestamp) < config.Cfg.Feature.SNI.TimeWindow {
			result = append(result, f)
		}
	}
	return result
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
