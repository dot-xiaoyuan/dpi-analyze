package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"
)

var Mongo mongodb
var Context context.Context

func Setup() error {
	return Mongo.Setup()
}

type mongodb struct {
	once        sync.Once
	initialized bool
	client      *mongo.Client
}

func (m *mongodb) Setup() error {
	var setupErr error
	m.once.Do(func() {
		if m.initialized {
			setupErr = fmt.Errorf("mongo database already initialized")
			return
		}

		if config.Cfg.Mongodb.Host == "" {
			setupErr = fmt.Errorf("mongodb.host is empty")
			return
		}
		if config.Cfg.Mongodb.Port == "" {
			setupErr = fmt.Errorf("mongodb.port is empty")
			return
		}

		uri := fmt.Sprintf("mongodb://%s:%s", config.Cfg.Mongodb.Host, config.Cfg.Mongodb.Port)

		opts := options.Client().
			ApplyURI(uri).
			SetServerSelectionTimeout(3 * time.Second)

		var err error
		Context = context.Background()
		m.client, err = mongo.Connect(Context, opts)
		if err != nil {
			zap.L().Error("Failed connecting to mongodb", zap.Error(err))
			setupErr = err
			return
		}
		err = m.client.Ping(Context, nil)
		if err != nil {
			zap.L().Error("Failed pinging mongodb", zap.Error(err))
			setupErr = err
			return
		}

		zap.L().Info("Connected to mongodb", zap.String("uri", uri))
		m.initialized = true
		// 加载配置
		err = searchMongo()
		if err != nil {
			zap.L().Error("Failed initializing mongodb", zap.Error(err))
			setupErr = err
			return
		}
	})
	return setupErr
}

// InsertOneStream 插入集合
func (m *mongodb) InsertOneStream(c string, d interface{}) error {
	if m.client == nil {
		zap.L().Error("Mongodb Client Not Initialized")
		return fmt.Errorf("mongodb Client Not Initialized")
	}

	collection := m.client.Database(types.MongoDatabaseStream).Collection(time.Now().Format(c + "-06-01-02-15"))
	_, err := collection.InsertOne(Context, d)
	if err != nil {
		zap.L().Error("Mongodb InsertOneStream Error", zap.Error(err))
	}
	return err
}

func GetMongoClient() *mongo.Client {
	if Mongo.client == nil {
		if err := Mongo.Setup(); err != nil {
			return nil
		}
	}
	return Mongo.client
}

// 查找mongodb的配置
func searchMongo() error {
	collection := GetMongoClient().Database(types.MongoDatabaseConfigs).Collection(types.MongoCollectionConfig)

	if count, err := collection.CountDocuments(Context, bson.M{}); count == 0 || err != nil {
		// 配置不存在，将yaml加载到mongodb中
		return store2Mongo()
	}
	// 配置存在，将mongodb中的配置重载到config
	_, err := collection.Find(Context, bson.M{})
	if err != nil {
		zap.L().Fatal("Failed to Find Mongodb Config", zap.Error(err))
		return err
	}

	var configDoc config.Yaml
	err = collection.FindOne(Context, bson.M{}).Decode(&configDoc)
	if err != nil {
		zap.L().Fatal("Failed to Find Mongodb Config", zap.Error(err))
		return err
	}

	// 将 MongoDB 配置加载到 config.Cfg
	config.Cfg = &configDoc
	return nil
}

// 将配置保存到mongodb中
func store2Mongo() error {
	collection := GetMongoClient().Database(types.MongoDatabaseConfigs).Collection(types.MongoCollectionConfig)

	_, err := collection.UpdateOne(
		Context,
		bson.M{"_id": "runtime_config"},
		bson.M{"$set": config.Cfg},
		options.Update().SetUpsert(true))
	if err != nil {
		zap.L().Panic("Mongodb UpdateOne Error", zap.Error(err))
		return err
	}
	return nil
}

func UpdateConfig(p config.Yaml) error {
	zap.L().Debug("UpdateConfig", zap.Any("config", p))

	//if len(p.LogLevel) > 0 && p.LogLevel != config.Cfg.LogLevel {
	//	config.Cfg.LogLevel = p.LogLevel
	//}
	//if p.Debug != config.Cfg.Debug {
	//	config.Cfg.Debug = p.Debug
	//}
	//if p.IgnoreMissing != config.Cfg.IgnoreMissing {
	//	config.Cfg.IgnoreMissing = p.IgnoreMissing
	//}
	//if p.FollowOnlyOnlineUsers != config.Cfg.FollowOnlyOnlineUsers {
	//	config.Cfg.FollowOnlyOnlineUsers = p.FollowOnlyOnlineUsers
	//}
	//if p.UseTTL != config.Cfg.UseTTL {
	//	config.Cfg.UseTTL = p.UseTTL
	//}
	//if p.UseUA != config.Cfg.UseUA {
	//	config.Cfg.UseUA = p.UseUA
	//}
	//if p.UseFeature != config.Cfg.UseFeature {
	//	config.Cfg.UseFeature = p.UseFeature
	//}

	return MergeStruct(config.Cfg, &p)
	//return nil
}

func MergeStruct(dst, src interface{}) error {
	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)

	// 确保目标是指针类型且非空
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return errors.New("destination must be a non-nil pointer")
	}

	dstVal = dstVal.Elem()
	srcVal = srcVal.Elem()

	// 确保目标和源是结构体
	if dstVal.Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		return errors.New("both destination and source must be struct pointers")
	}

	// 遍历目标字段
	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)         // 目标结构体的字段值
		srcField := srcVal.Field(i)         // 来源结构体的字段值
		fieldType := dstVal.Type().Field(i) // 字段元信息

		// 如果字段是嵌套结构体，递归合并
		if dstField.Kind() == reflect.Struct && srcField.Kind() == reflect.Struct {
			err := MergeStruct(dstField.Addr().Interface(), srcField.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}

		// 如果字段是零值，跳过更新
		if isZeroValue(srcField) {
			continue
		}
		// 如果字段值不同，更新目标字段
		if !reflect.DeepEqual(dstField.Interface(), srcField.Interface()) {
			if dstField.CanSet() {
				dstField.Set(srcField)
				zap.L().Debug("set", zap.Any("dstField", dstField.Interface()))
				zap.L().Debug("debug", zap.Bool("d", config.Cfg.Debug), zap.Any("ft", fieldType))
			} else {
				return errors.New("cannot set value for field: " + fieldType.Name)
			}
		}
	}

	return nil
}

// 判断字段是否是零值
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
