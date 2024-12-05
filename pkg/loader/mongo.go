package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
	"time"
)

type MongoLoader struct {
	Client             *mongo.Client
	MetadataCollection string
	HistoryCollection  string
	Database           string
	Version            string
}

// Load 获取当前最新版本的数据
func (ml *MongoLoader) Load() ([]byte, error) {
	ctx := context.TODO()

	metadataColl := ml.Client.Database(ml.Database).Collection(ml.MetadataCollection)
	historyColl := ml.Client.Database(ml.Database).Collection(ml.HistoryCollection)

	// 获取当前元数据中的版本号
	var metadata struct {
		CurrentVersion string `bson:"current_version"`
	}
	err := metadataColl.FindOne(ctx, bson.M{}).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	// 根据版本号在历史集合中获取对应数据
	var history struct {
		Data []byte `bson:"data"`
	}
	err = historyColl.FindOne(ctx, bson.M{"version": metadata.CurrentVersion}).Decode(&history)
	if err != nil {
		return nil, err
	}
	return history.Data, nil
}

func (ml *MongoLoader) Exists() bool {
	exists, err := mongodb.CollectionExists(ml.Database, ml.MetadataCollection)
	if err != nil {
		return false
	}
	return exists
}

// Save 存储新版本数据
func (ml *MongoLoader) Save(rawData []byte) error {
	ctx := context.TODO()
	// 获取版本
	line, _ := bytes.NewBuffer(rawData).ReadBytes('\n')
	re := regexp.MustCompile(`v\d{2}\.\d{2}\.\d{2}`)
	if strings.Contains(string(line), "version") {
		matches := re.FindString(string(line))
		ml.Version = strings.Trim(matches, "\n")
	}
	metadataColl := ml.Client.Database(ml.Database).Collection(ml.MetadataCollection)
	historyColl := ml.Client.Database(ml.Database).Collection(ml.HistoryCollection)
	// 获取当前元信息
	var metadata struct {
		CurrentVersion string `bson:"current_version"`
	}
	docID := primitive.NewObjectID()
	err := metadataColl.FindOne(ctx, bson.M{"_id": docID}).Decode(&metadata)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 如果元数据不存在，则初始化
			_, err = metadataColl.InsertOne(ctx, bson.M{
				"_id":             docID,
				"current_version": ml.Version,
				"created_at":      time.Now(),
				"updated_at":      time.Now(),
			})
			if err != nil {
				return err
			}
			// 在历史集合中插入第一版数据
			_, err = historyColl.InsertOne(ctx, bson.M{
				"doc_id":    docID,
				"version":   ml.Version,
				"data":      rawData,
				"timestamp": time.Now(),
			})
			return err
		}
		return err
	}

	// 计算新版本号
	newVersion := time.Now().Format("v06.01.02")

	// 更新元集合中的版本信息
	_, err = metadataColl.UpdateOne(ctx, bson.M{"_id": docID}, bson.M{
		"$set": bson.M{
			"current_version": newVersion,
			"updated_at":      time.Now(),
		},
	})
	if err != nil {
		return err
	}

	// 在历史集合中插入新版本数据
	_, err = historyColl.InsertOne(ctx, bson.M{
		"doc_id":    docID,
		"version":   newVersion,
		"data":      rawData,
		"timestamp": time.Now(),
	})
	return err
}

// GetCurrentVersion 查询当前版本
func (ml *MongoLoader) GetCurrentVersion() (string, error) {
	collection := ml.Client.Database(ml.Database).Collection(ml.MetadataCollection)

	// 查询最新的版本，假设版本字段是 "version" 并按时间排序
	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}}) // 按创建时间降序
	filter := bson.D{}                                            // 可扩展为按条件查询
	var result struct {
		Version string `bson:"current_version"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", fmt.Errorf("no version found in collection %s", ml.MetadataCollection)
		}
		return "", err
	}

	return result.Version, nil
}

// RollbackToVersion 回滚数据到指定版本
func RollbackToVersion(metadataColl *mongo.Collection, historyColl *mongo.Collection, docID primitive.ObjectID, version int32) error {
	ctx := context.TODO()

	// 在历史集合中找到指定版本的数据
	var history struct {
		Data []byte `bson:"data"`
	}
	err := historyColl.FindOne(ctx, bson.M{"doc_id": docID, "version": version}).Decode(&history)
	if err != nil {
		return err
	}

	// 更新元集合中的版本号和更新时间
	_, err = metadataColl.UpdateOne(ctx, bson.M{"_id": docID}, bson.M{
		"$set": bson.M{
			"current_version": version,
			"updated_at":      time.Now(),
		},
	})
	return err
}
