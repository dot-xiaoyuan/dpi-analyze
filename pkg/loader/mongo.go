package loader

import (
	"context"
	"fmt"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoLoader struct {
	Client     *mongo.Client
	Collection string
	Database   string
}

func (ml *MongoLoader) Load() ([]byte, error) {
	cursor, err := ml.Client.Database(ml.Database).Collection(ml.Collection).Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	for cursor.Next(context.TODO()) {
		var record bson.M
		err = cursor.Decode(&record)
		if err != nil {
			return nil, err
		}
		results = append(results, record)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return bson.Marshal(results)
}

func (ml *MongoLoader) Exists() bool {
	exists, err := mongodb.CollectionExists(ml.Database, ml.Collection)
	if err != nil {
		return false
	}
	return exists
}

func (ml *MongoLoader) Save(data []byte) error {
	// 插入到 MongoDB
	collection := ml.Client.Database(ml.Database).Collection(ml.Collection)
	_, err := collection.InsertOne(context.TODO(), bson.D{
		{"data", data}, // 存储为二进制数据
	})
	if err != nil {
		return fmt.Errorf("failed to insert data into MongoDB: %w", err)
	}

	return nil
}
