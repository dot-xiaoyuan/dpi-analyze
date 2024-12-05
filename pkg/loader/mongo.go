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

type History struct {
	DocId        primitive.ObjectID `bson:"doc_id" json:"doc_id"`
	Version      string             `bson:"version" json:"version"`
	Type         string             `bson:"type" json:"type"`
	Timestamp    primitive.DateTime `bson:"timestamp" json:"timestamp"`
	ChangeNumber int                `bson:"change_number" json:"change_number"`
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
func (ml *MongoLoader) Save(rawData []byte, changeNumber int) error {
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
		ID             primitive.ObjectID `bson:"_id"`
		CurrentVersion string             `bson:"current_version"`
		CreatedAt      primitive.DateTime `bson:"created_at"`
		UpdatedAt      primitive.DateTime `bson:"updated_at"`
	}
	docID := primitive.NewObjectID()
	err := metadataColl.FindOne(ctx, bson.M{}).Decode(&metadata)
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
				"doc_id":        docID,
				"version":       ml.Version,
				"data":          rawData,
				"timestamp":     time.Now(),
				"change_number": 0,
				"type":          "insert",
			})
			return err
		}
		return err
	}

	// 计算新版本号
	//newVersion := time.Now().Format("v06.01.02")

	// 更新元集合中的版本信息
	_, err = metadataColl.UpdateOne(ctx, bson.M{"_id": metadata.ID}, bson.M{
		"$set": bson.M{
			"current_version": ml.Version,
			"updated_at":      time.Now(),
		},
	})
	if err != nil {
		return err
	}

	// 在历史集合中插入新版本数据
	_, err = historyColl.InsertOne(ctx, bson.M{
		"doc_id":        docID,
		"version":       ml.Version,
		"data":          rawData,
		"timestamp":     time.Now(),
		"change_number": changeNumber,
		"type":          "update",
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

// GetHistoryVersions 获取历史版本信息
func (ml *MongoLoader) GetHistoryVersions() ([]History, error) {
	collection := ml.Client.Database(ml.Database).Collection(ml.HistoryCollection)

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 按创建时间降序
	filter := bson.D{}                                         // 可扩展为按条件查询

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行查询
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("no version found in collection %s", ml.HistoryCollection)
		}
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []History

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
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
