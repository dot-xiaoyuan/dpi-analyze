package cmd

import (
	"context"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清空所有数据",
	Run: func(cmd *cobra.Command, args []string) {
		// 连接 MongoDB
		if err := mongo.Setup(); err != nil {
			fmt.Printf("🚨 mongo 连接失败: %v\n", err)
			os.Exit(1)
		}

		// 初始化 spinner
		spinner := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
		spinner.Start()

		// 需要清空的数据库表
		tables := []string{
			types.MongoDatabaseStream,
			types.MongoDatabaseDevices,
			types.MongoDatabaseFeatures,
			types.MongoDatabaseProxy,
			types.MongoDatabaseUserEvents,
			types.MongoDatabaseUserAgent,
			types.MongoDatabaseSuspected,
		}

		// 清空 MongoDB 操作
		for _, table := range tables {
			time.Sleep(time.Second / 2)
			// 停止 spinner，并打印当前清空的表格信息
			spinner.Stop()
			fmt.Printf("🔄 正在清空 %s 表...\n", table)

			// 执行删除操作
			err := mongo.GetMongoClient().Database(table).Drop(mongo.Context)
			if err != nil {
				// 如果删除失败，打印失败信息
				fmt.Printf("❌ 清空 %s 表失败: %v\n", table, err)
			} else {
				// 删除成功，打印成功信息
				fmt.Printf("✅ 已清空 %s 表\n", table)
			}

			// 重新开始 spinner
			spinner.Start()
		}

		// 需要清空的 Redis 键集合
		setKeys := []string{
			types.ZSetIP,
			types.ZSetApplication,
			types.ZSetOnlineUsers,
			types.ZSetObserverTTL,
			types.ZSetObserverMac,
			types.ZSetObserverUa,
			types.ZSetObserverDevice,
		}

		// 获取 Redis 客户端
		rdb := redis.GetRedisClient()
		ctx := context.TODO()

		// 清空 Redis 键集合
		for _, key := range setKeys {
			time.Sleep(time.Second / 2)

			// 停止 spinner，并打印当前清空的集合信息
			spinner.Stop()
			fmt.Printf("🔄 正在清空 %s 集合...\n", key)

			// 执行删除操作
			err := rdb.Del(ctx, key).Err()
			if err != nil {
				// 如果删除失败，打印失败信息
				fmt.Printf("❌ 清空 %s 集合失败: %v\n", key, err)
			} else {
				// 删除成功，打印成功信息
				fmt.Printf("✅ 已清空 %s 集合\n", key)
			}

			// 重新开始 spinner
			spinner.Start()
		}

		// 使用 SCAN 扫描并删除键
		scanKeys := []string{
			types.HashAnalyzeIP,
			types.SetIPDevices,
			types.KeyDiscoverIP,
			types.KeyDevicesMobileIP,
			types.KeyDevicesPcIP,
			types.KeyDevicesAllIP,
		}

		// 优化 SCAN 操作
		for _, key := range scanKeys {
			runes := []rune(key)
			prefix := string(runes[0:len(runes)-2]) + "*"
			var cursor uint64
			var keysToDelete []string

			// 使用 SCAN 遍历所有键并删除
			for {
				// 停止 spinner，提示扫描当前前缀
				spinner.Stop()
				fmt.Printf("🔄 正在扫描并删除 %s 前缀的键...\n", prefix)

				// 使用 SCAN 获取匹配前缀的键
				var keys []string
				var err error
				keys, cursor, err = rdb.Scan(ctx, cursor, prefix, 100).Result() // 批量返回更多键，减少请求次数
				if err != nil {
					// 如果扫描失败，打印错误信息
					fmt.Printf("❌ 扫描失败: %v\n", err)
					return
				}

				// 将键添加到删除列表
				keysToDelete = append(keysToDelete, keys...)

				// 如果 cursor 为 0，表示扫描结束
				if cursor == 0 {
					break
				}
			}

			// 如果有键需要删除，批量删除
			if len(keysToDelete) > 0 {
				// 停止 spinner，提示删除操作
				spinner.Stop()
				fmt.Printf("🔄 正在删除 %d 个键...\n", len(keysToDelete))

				// 执行删除操作
				_, err := rdb.Del(ctx, keysToDelete...).Result()
				if err != nil {
					// 如果删除失败，打印错误信息
					fmt.Printf("❌ 删除 %d 个键失败: %v\n", len(keysToDelete), err)
				} else {
					// 删除成功，打印成功信息
					fmt.Printf("✅ 已删除 %d 个键\n", len(keysToDelete))
				}
			} else {
				fmt.Println("没有找到匹配的键")
			}

			// 重新开始 spinner
			spinner.Start()
		}

		// 清空完成后，停止 spinner
		spinner.Stop()
		fmt.Println("\n🎉 所有表格清空操作完成！")
	},
}
