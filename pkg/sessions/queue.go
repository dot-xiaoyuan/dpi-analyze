package sessions

import (
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
	"time"
)

var SessionQueue = make(chan types.Sessions, 500000)

func StartLogConsumer() {
	consumerCount := 4 // 启动 4 个消费者，你可以根据机器的资源来调整这个值
	for i := 0; i < consumerCount; i++ {
		go consumeLogQueue()
	}
}

func consumeLogQueue() {
	buffer := make([]interface{}, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case log := <-SessionQueue:
			buffer = append(buffer, log)
			if len(buffer) >= 1000 {
				insertManyStream(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				insertManyStream(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func insertManyStream(buffer []interface{}) {
	retries := 3
	var err error
	for i := 0; i < retries; i++ {
		_, err = mongodb.GetMongoClient().Database(types.MongoDatabaseStream).Collection(time.Now().Format("stream-06-01-02-15")).
			InsertMany(mongodb.Context, buffer)
		if err == nil {
			return // 成功插入
		}
		time.Sleep(time.Duration(1<<i) * time.Second) // 指数退避
	}
	// 记录失败
	zap.L().Error("insert stream failed after retries", zap.Error(err))
}
