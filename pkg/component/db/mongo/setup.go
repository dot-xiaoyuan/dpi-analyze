package mongo

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"sync"
	"time"
)

var Mongo mongodb

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
		m.client, err = mongo.Connect(context.TODO(), opts)
		if err != nil {
			zap.L().Error("Failed connecting to mongodb", zap.Error(err))
			setupErr = err
			return
		}
		err = m.client.Ping(context.TODO(), nil)
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

// InsertOne 插入集合
func (m *mongodb) InsertOne(c string, d interface{}) error {
	if m.client == nil {
		zap.L().Error("Mongodb Client Not Initialized")
		return fmt.Errorf("mongodb Client Not Initialized")
	}

	collection := m.client.Database("dpi").Collection(time.Now().Format(c + "-06-01-02-15"))
	_, err := collection.InsertOne(context.TODO(), d)
	if err != nil {
		zap.L().Error("Mongodb InsertOne Error", zap.Error(err))
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
