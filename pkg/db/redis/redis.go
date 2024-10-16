package redis

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"sync"
)

var (
	one    sync.Once
	Client *redis.Client
	Online *redis.Client
	Cache  *redis.Client
)

func Setup() {
	one.Do(func() {
		var err error
		// setup dpi client
		Client, err = loadRedisClient(config.Cfg.Redis.DPI)
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
		// setup online client
		Online, err = loadRedisClient(config.Cfg.Redis.Online)
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
		// setup cache client
		Cache, err = loadRedisClient(config.Cfg.Redis.Cache)
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	})
}

func loadRedisClient(c config.RedisConfig) (*redis.Client, error) {
	if c.Host == "" {
		return nil, fmt.Errorf("redis host is empty")
	}
	if c.Port == "" {
		return nil, fmt.Errorf("redis port is empty")
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
	})

	// Ping Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		zap.L().Error(i18n.T("Failed to ping Redis"))
		return nil, err
	} else {
		zap.L().Info(i18n.TT("Connected to Redis!", map[string]interface{}{
			"host": c.Host,
			"port": c.Port,
		}))
		return rdb, nil
	}
}

func GetRedisClient() *redis.Client {
	if Client == nil {
		Setup()
	}
	return Client
}

func GetOnlineRedisClient() *redis.Client {
	if Online == nil {
		Setup()
	}
	return Online
}

func GetCacheRedisClient() *redis.Client {
	if Cache == nil {
		Setup()
	}
	return Cache
}
