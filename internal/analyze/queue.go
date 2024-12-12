package analyze

import (
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.uber.org/zap"
	"time"
)

var logQueue = make(chan types.Sessions, 10000)

func StartLogConsumer() {
	go func() {
		buffer := make([]any, 0, 100)
		ticker := time.NewTicker(time.Second) // 每秒批量写入
		defer ticker.Stop()

		for {
			select {
			case log := <-logQueue:
				buffer = append(buffer, log)
				if len(buffer) >= 100 {
					insertManyStream(buffer)
					buffer = buffer[:0]
				}
			case <-ticker.C: // 定时写入
				if len(buffer) > 0 {
					insertManyStream(buffer)
					buffer = buffer[:0]
				}
			}
		}
	}()
}

func insertManyStream(buffer []interface{}) {
	_, err := mongodb.GetMongoClient().Database(types.MongoDatabaseStream).Collection(time.Now().Format("stream-06-01-02-15")).
		InsertMany(mongodb.Context, buffer)
	if err != nil {
		zap.L().Error("insert stream failed", zap.Error(err))
		return
	}
}
