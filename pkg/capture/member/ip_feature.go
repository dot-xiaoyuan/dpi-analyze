package member

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"log"
	"sync"
	"time"
)

// 特征

type Feature struct {
	IP    string
	Field types.Feature
	Value string
}

var (
	featureCaches = make(map[string]*types.FeatureSet) // IP为key的特征集合缓存
	cacheLock     sync.RWMutex                         // 缓存锁
)

// GetFeatureSet 获取或创建IP对应的FeatureSet
func GetFeatureSet(ip string) *types.FeatureSet {
	cacheLock.RLock()
	featureSet, exists := featureCaches[ip]
	cacheLock.RUnlock()

	if !exists {
		cacheLock.Lock()
		defer cacheLock.Unlock()
		featureSet = &types.FeatureSet{
			LastSeen: time.Now(),
			Features: make(map[types.Feature][]types.FeatureData),
		}
		featureCaches[ip] = featureSet
	}
	return featureSet
}

// Increment 增量统计特征访问频率
func Increment(f Feature) {

	featureSet := GetFeatureSet(f.IP)

	cacheLock.Lock()
	defer cacheLock.Unlock()

	// 获取当前特征的列表
	featureList, exists := featureSet.Features[f.Field]
	if !exists {
		// 如果该字段没有数据，则初始化列表
		featureSet.Features[f.Field] = []types.FeatureData{
			{LastSeen: time.Now(), Value: f.Value, Count: 1},
		}
		featureSet.Total = append(featureSet.Total, types.Chart{
			Date:       time.Now(),
			Industry:   f.Field,
			Unemployed: 1,
		})
		return
	}

	// 遍历当前特征列表，检查是否存在相同的 Value
	for i, feature := range featureList {
		if feature.Value == f.Value {
			// 如果找到相同的 Value，则更新 LastSeen 时间
			featureList[i].LastSeen = time.Now()
			featureList[i].Count++

			// 更新 Total 中的 Chart 数据
			updateChart(featureSet, f.Field, featureList[i].Count)
			return
		}
	}

	// 如果没有相同的 Value，则追加新的特征数据
	featureSet.Features[f.Field] = append(
		featureList,
		types.FeatureData{LastSeen: time.Now(), Value: f.Value, Count: 1},
	)
}

// FlushToMongo 刷新到mongodb
func FlushToMongo() {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	// 遍历缓存并插入MongoDB
	for ip, featureSet := range featureCaches {
		_, err := mongo.GetMongoClient().Database("features").Collection(fmt.Sprintf("ip-%s", ip)).InsertOne(context.TODO(), featureSet)
		if err != nil {
			log.Printf("Failed to insert data for IP %s: %v", ip, err)
		}
		delete(featureCaches, ip) // 清空缓存
	}
}

// StartFlushScheduler 刷新时间窗口，记录到mongodb
func StartFlushScheduler(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		FlushToMongo()
	}
}

// updateChart 更新 Total 切片中的 Chart 数据
func updateChart(featureSet *types.FeatureSet, industry types.Feature, newCount int) {
	for i, chart := range featureSet.Total {
		if chart.Industry == industry {
			// 更新相应的 Chart 数据
			featureSet.Total[i].Unemployed = newCount
			return
		}
	}
	// 如果没有找到相应的 Chart，则添加新的
	featureSet.Total = append(featureSet.Total, types.Chart{
		Date:       time.Now(),
		Industry:   industry,
		Unemployed: newCount,
	})
}
