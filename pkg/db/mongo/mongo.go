package mongo

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

var (
	client   *mongo.Client
	one      sync.Once
	mongoURI string
)

func Setup() error {
	var setupErr error
	one.Do(func() {
		err := loadMongoClient()
		if err != nil {
			zap.L().Error(err.Error())
			setupErr = err
			return
		}
	})
	return setupErr
}

/**
 * 加载mongodb客户端
 */
func loadMongoClient() (err error) {
	if config.Cfg.Mongodb.Host == "" {
		return fmt.Errorf(i18n.T("mongodb.host is empty"))
	}
	if config.Cfg.Mongodb.Port == "" {
		return fmt.Errorf(i18n.T("mongodb.port is empty"))
	}
	mongoURI = fmt.Sprintf("mongodb://%s:%s", config.Cfg.Mongodb.Host, config.Cfg.Mongodb.Port)

	opt := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(3 * time.Second)

	client, err = mongo.Connect(
		context.TODO(),
		opt,
	)
	if err != nil {
		return fmt.Errorf(i18n.T("Failed to connect to MongoDB"))
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf(i18n.T("Failed to ping MongoDB"))
	}
	zap.L().Info(i18n.TT("Connected to MongoDB!", map[string]interface{}{
		"host": config.Cfg.Mongodb.Host,
		"port": config.Cfg.Mongodb.Port,
	}))
	return
}

// GetMongoClient 获取mongo实例
func GetMongoClient() *mongo.Client {
	if client == nil {
		err := loadMongoClient()
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	}
	return client
}

// InsertOne 插入集合
func InsertOne(collectionName string, document interface{}) error {
	c := GetMongoClient()

	if client == nil {
		zap.L().Error(i18n.T("MongoDB client not initialized"))
		os.Exit(1)
	}

	collection := c.Database("dpi").Collection(time.Now().Format(collectionName + "-06-01-02-15"))
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}
