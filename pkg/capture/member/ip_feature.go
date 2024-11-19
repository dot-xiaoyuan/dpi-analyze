package member

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 特征
var (
	featureCaches = make(map[string]*types.FeatureSet) // IP为key的特征集合缓存
	cacheLock     sync.RWMutex                         // 缓存锁
	indexOnce     sync.Once
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
			IP:       ip,
			LastSeen: time.Now(),
			Features: make(map[types.FeatureType][]types.FeatureData),
		}
		featureCaches[ip] = featureSet
	}
	return featureSet
}

func DelFeatureSet(ip string) {
	cacheLock.RLock()
	_, exists := featureCaches[ip]
	cacheLock.RUnlock()

	if exists {
		delete(featureCaches, ip)
	}
}

// Increment 增量统计特征访问频率
func Increment(f types.Feature) {
	featureSet := GetFeatureSet(f.IP)

	cacheLock.Lock()
	defer cacheLock.Unlock()

	now := time.Now()
	// 获取当前特征的列表
	featureList, exists := featureSet.Features[f.Field]
	if !exists {
		// 如果该字段没有数据，则初始化列表
		featureSet.Features[f.Field] = []types.FeatureData{
			{LastSeen: now, Value: f.Value, Count: 1},
		}
		featureSet.Total = append(featureSet.Total, types.Chart{
			Date:       featureSet.LastSeen,
			Industry:   f.Field,
			Unemployed: 1,
		})
		return
	}

	// 遍历当前特征列表，检查是否存在相同的 Value
	for i, feature := range featureList {
		if feature.Value == f.Value {
			// 如果找到相同的 Value，则更新 LastSeen 时间
			featureList[i].LastSeen = now
			featureList[i].Count++

			// 更新 Total 中的 Chart 数据
			updateChart(featureSet, f.Field, featureList[i].Count)
			return
		}
	}

	// 如果没有相同的 Value，则追加新的特征数据
	featureSet.Features[f.Field] = append(
		featureList,
		types.FeatureData{LastSeen: now, Value: f.Value, Count: 1},
	)
}

// FlushToMongo 刷新到mongodb
func FlushToMongo() {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	// 遍历缓存并插入MongoDB
	var docs []interface{}
	for ip, featureSet := range featureCaches {
		docs = append(docs, featureSet)
		delete(featureCaches, ip) // 清空缓存
	}
	err := batchInsertToMongo(docs)
	if err != nil {
		zap.L().Error("batch insert to mongo failed", zap.Error(err))
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
func updateChart(featureSet *types.FeatureSet, industry types.FeatureType, newCount int) {
	for i, chart := range featureSet.Total {
		if chart.Industry == industry {
			// 更新相应的 Chart 数据
			featureSet.Total[i].Unemployed = newCount
			return
		}
	}
	// 如果没有找到相应的 Chart，则添加新的
	featureSet.Total = append(featureSet.Total, types.Chart{
		Date:       featureSet.LastSeen,
		Industry:   industry,
		Unemployed: newCount,
	})
}

// TTL索引
func ensureIndex() error {
	collection := mongo.GetMongoClient().Database(types.MongoDatabaseFeatures).Collection(types.OnlineUsersFeature)
	_, err :=
		collection.Indexes().CreateMany(
			mongo.Context,
			[]mgo.IndexModel{
				{
					Keys: bson.D{{"ip", 1}},
				},
				{
					Keys:    bson.D{{"last_seen", 1}},
					Options: options.Index().SetExpireAfterSeconds(60 * 30),
				},
			},
		)
	return err
}

// 批量插入到mongodb
func batchInsertToMongo(docs []interface{}) error {
	collection := mongo.GetMongoClient().Database(types.MongoDatabaseFeatures).Collection(types.OnlineUsersFeature)

	batchSize := 500
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}

		// 批量插入
		_, err := collection.InsertMany(mongo.Context, docs[i:end])
		if err != nil {
			zap.L().Error("Batch insert to mongo failed", zap.Error(err))
			return err
		}
	}
	return nil
}

// EnsureIndexOnce 设置索引
func EnsureIndexOnce() {
	indexOnce.Do(func() {
		if err := ensureIndex(); err != nil {
			zap.L().Panic("EnsureIndex failed", zap.Error(err))
		}
	})
}
