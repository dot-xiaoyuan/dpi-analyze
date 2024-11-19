package cmd

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清空所有数据",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mongo.Setup(); err != nil {
			zap.L().Error("mongo 连接失败", zap.Error(err))
			os.Exit(1)
		}
		var err error
		err = mongo.GetMongoClient().Database(types.MongoDatabaseStream).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空流日志失败", zap.Error(err))
		} else {
			zap.L().Info("已清空流日志")
		}
		err = mongo.GetMongoClient().Database(types.MongoDatabaseDevices).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空设备表失败", zap.Error(err))
		} else {
			zap.L().Info("已清空设备表")
		}
		err = mongo.GetMongoClient().Database(types.MongoDatabaseFeatures).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空流量特征表失败", zap.Error(err))
		} else {
			zap.L().Info("已清空流量特征表")
		}
		err = mongo.GetMongoClient().Database(types.MongoDatabaseProxy).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空代理记录表失败", zap.Error(err))
		} else {
			zap.L().Info("已清空代理记录表")
		}
		err = mongo.GetMongoClient().Database(types.MongoDatabaseEvents).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空上下线日志表失败", zap.Error(err))
		} else {
			zap.L().Info("已清空上下线日志表")
		}
		err = mongo.GetMongoClient().Database(types.MongoDatabaseUserAgent).Drop(mongo.Context)
		if err != nil {
			zap.L().Error("清空useragent表失败", zap.Error(err))
		} else {
			zap.L().Info("已清空useragent表")
		}
	},
}
