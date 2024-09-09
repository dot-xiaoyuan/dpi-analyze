package mongo

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
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

func Setup() {
	one.Do(func() {
		err := loadMongoClient()
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	})
}

func loadMongoClient() (err error) {
	spinners.Start()
	defer func() {
		spinners.Stop()
	}()
	if config.Cfg.Mongodb.Host == "" {
		return fmt.Errorf(i18n.Translate.T("mongodb.host is empty", nil))
	}
	if config.Cfg.Mongodb.Port == "" {
		return fmt.Errorf(i18n.Translate.T("mongodb.port is empty", nil))
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
		return fmt.Errorf(i18n.Translate.T("Failed to connect to MongoDB", nil))
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf(i18n.Translate.T("Failed to ping MongoDB", nil))
	}
	zap.L().Info(i18n.Translate.T("Connected to MongoDB!", nil))
	return
}

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

func InsertOne(collectionName string, document interface{}) error {
	c := GetMongoClient()

	if client == nil {
		zap.L().Error(i18n.Translate.T("MongoDB client not initialized", nil))
		os.Exit(1)
	}

	collection := c.Database("dpi").Collection(time.Now().Format(collectionName + "-06-01-02-15"))
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}
