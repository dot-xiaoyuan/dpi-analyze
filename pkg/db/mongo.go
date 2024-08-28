package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
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

func GetMongoClient() *mongo.Client {
	one.Do(func() {
		var err error
		if config.Cfg.Mongodb.Host == "" {
			panic(errors.New("mongodb.host is empty"))
		}
		if config.Cfg.Mongodb.Port == "" {
			panic(errors.New("mongodb.port is empty"))
		}
		mongoURI = fmt.Sprintf("mongodb://%s:%s", config.Cfg.Mongodb.Host, config.Cfg.Mongodb.Port)
		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
		if err != nil {
			zap.L().Error("Failed to connect to MongoDB", zap.Error(err))
			os.Exit(1)
		}

		err = client.Ping(context.TODO(), nil)
		if err != nil {
			zap.L().Error("Failed to ping MongoDB", zap.Error(err))
			os.Exit(1)
		}
		zap.L().Info("Connected to MongoDB!")
	})

	return client
}

func InsertOne(collectionName string, document interface{}) error {
	c := GetMongoClient()

	if client == nil {
		panic(errors.New("MongoDB client not initialized"))
	}

	collection := c.Database("dpi").Collection(time.Now().Format(collectionName + "_20060102_15"))
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}
