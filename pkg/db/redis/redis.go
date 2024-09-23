package redis

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"sync"
)

var (
	one         sync.Once
	RedisClient *redis.Client
)

func Setup() {
	one.Do(func() {
		err := loadRedisClient()
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	})
}

func loadRedisClient() error {
	spinners.Start()
	defer func() {
		spinners.Stop()
	}()

	if config.Cfg.Redis.Host == "" {
		return fmt.Errorf(i18n.T("redis.host is empty"))
	}
	if config.Cfg.Redis.Port == "" {
		return fmt.Errorf(i18n.T("redis.port is empty"))
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Cfg.Redis.Host, config.Cfg.Redis.Port),
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})

	// Ping Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		zap.L().Error(i18n.T("Failed to ping Redis"))
		return err
	} else {
		zap.L().Info(i18n.TT("Connected to Redis!", map[string]interface{}{
			"host": config.Cfg.Redis.Host,
			"port": config.Cfg.Redis.Port,
		}))
		RedisClient = rdb
		return nil
	}
}

func GetRedisClient() *redis.Client {
	if RedisClient == nil {
		err := loadRedisClient()
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	}
	return RedisClient
}
