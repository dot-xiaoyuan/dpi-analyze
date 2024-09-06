package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
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
		loadMongoClient()
	})
}

func loadMongoClient() {
	var err error
	if config.Cfg.Mongodb.Host == "" {
		panic(errors.New(i18n.Translate.T("mongodb.host is empty", nil)))
	}
	if config.Cfg.Mongodb.Port == "" {
		panic(errors.New(i18n.Translate.T("mongodb.port is empty", nil)))
	}
	mongoURI = fmt.Sprintf("mongodb://%s:%s", config.Cfg.Mongodb.Host, config.Cfg.Mongodb.Port)
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		zap.L().Error(i18n.Translate.T("Failed to connect to MongoDB", nil), zap.Error(err))
		panic(errors.New(i18n.Translate.T("Failed to connect to MongoDB", nil)))
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		zap.L().Error(i18n.Translate.T("Failed to ping MongoDB", nil), zap.Error(err))
		panic(errors.New(i18n.Translate.T("Failed to ping MongoDB", nil)))
	}
	zap.L().Info(i18n.Translate.T("Connected to MongoDB!", nil))
}

func GetMongoClient() *mongo.Client {
	if client == nil {
		loadMongoClient()
	}
	return client
}

func InsertOne(collectionName string, document interface{}) error {
	c := GetMongoClient()

	if client == nil {
		panic(errors.New("MongoDB client not initialized"))
	}

	collection := c.Database("dpi").Collection(time.Now().Format(collectionName + "-06-01-02-15"))
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}
