package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

var (
	client   *mongo.Client
	once     sync.Once
	mongoURI string
	useMongo bool
)

func Setup(uri string, use bool) {
	mongoURI = uri
	useMongo = use
}

func GetMongoClient() *mongo.Client {
	once.Do(func() {
		if useMongo {
			var err error
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
		}
	})

	return client
}

func InsertOne(collectionName string, document interface{}) error {
	c := GetMongoClient()

	if client == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}

	collection := c.Database("dpi").Collection(time.Now().Format(collectionName + "20060102_15"))
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}
