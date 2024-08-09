package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"os"
)

// MongoDB

var MongoClient *mongo.Client

func Register(hostname, port string) {
	uri := fmt.Sprintf("mongodb://%s:%s", hostname, port)
	client, err := mongo.Connect(context.TODO(), options.Client().
		ApplyURI(uri))
	if err != nil {
		zap.L().Error("failed to connect to mongodb", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			zap.L().Error("failed to disconnect mongodb", zap.Error(err))
			os.Exit(1)
		}
	}()
	if err = client.Ping(context.TODO(), nil); err != nil {
		zap.L().Error("failed to ping mongodb", zap.Error(err))
		os.Exit(1)
	}
	MongoClient = client
}
