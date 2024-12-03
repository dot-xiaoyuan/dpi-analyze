package mongo

import (
	"context"
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
		return Store2Mongo()
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

// Store2Mongo 将配置保存到mongodb中
func Store2Mongo() error {
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

// UpdateNestedConfig 更新配置
func UpdateNestedConfig(cfg interface{}, updates map[string]interface{}) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")

		if newValue, ok := updates[jsonTag]; ok {
			if field.Kind() == reflect.Struct {
				// 递归更新嵌套结构
				if err := UpdateNestedConfig(field.Addr().Interface(), newValue.(map[string]interface{})); err != nil {
					return err
				}
			} else if field.CanSet() {
				newValueReflect := reflect.ValueOf(newValue)
				if newValueReflect.Type().ConvertibleTo(field.Type()) {
					field.Set(newValueReflect.Convert(field.Type()))
				} else {
					return fmt.Errorf("type mismatch for field %s", jsonTag)
				}
			}
		}
	}
	return nil
}

// CollectionExists 集合是否存在
func CollectionExists(database, collectionName string) (bool, error) {
	// 使用 ListCollections 检查集合是否存在
	filter := options.ListCollections().SetNameOnly(true)
	cursor, err := GetMongoClient().Database(database).ListCollections(Context, bson.M{"name": collectionName}, filter)
	if err != nil {
		return false, err
	}
	defer cursor.Close(Context)

	// 如果光标有结果，说明集合存在
	return cursor.Next(Context), nil
}
