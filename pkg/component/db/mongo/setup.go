package mongo

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
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
