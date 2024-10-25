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

type FeatureData struct {
	LastSeen time.Time `bson:"last_seen"` // 最后一次访问时间
	Value    string    `bson:"value"`     // 特征数值
	Count    int       `bson:"count"`     // 时间窗口内相同数值计数
}

type FeatureSet struct {
	LastSeen time.Time                       `bson:"last_seen"`
	Features map[types.Feature][]FeatureData `bson:"features"`
	Total    map[types.Feature][]int         `bson:"total"`
}

var (
	featureCaches = make(map[string]*FeatureSet) // IP为key的特征集合缓存
	cacheLock     sync.RWMutex                   // 缓存锁
)

// GetFeatureSet 获取或创建IP对应的FeatureSet
func GetFeatureSet(ip string) *FeatureSet {
	cacheLock.RLock()
	featureSet, exists := featureCaches[ip]
	cacheLock.RUnlock()

	if !exists {
		cacheLock.Lock()
		defer cacheLock.Unlock()
		featureSet = &FeatureSet{
			LastSeen: time.Now(),
			Features: make(map[types.Feature][]FeatureData),
			Total:    make(map[types.Feature][]int),
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
		featureSet.Features[f.Field] = []FeatureData{
			{LastSeen: time.Now(), Value: f.Value, Count: 1},
		}
		return
	}

	// 遍历当前特征列表，检查是否存在相同的 Value
	for i, feature := range featureList {
		if feature.Value == f.Value {
			// 如果找到相同的 Value，则更新 LastSeen 时间
			featureList[i].LastSeen = time.Now()
			featureList[i].Count++
			return
		}
	}

	// 如果没有相同的 Value，则追加新的特征数据
	featureSet.Features[f.Field] = append(
		featureList,
		FeatureData{LastSeen: time.Now(), Value: f.Value, Count: 1},
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
