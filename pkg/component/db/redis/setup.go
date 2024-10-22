package redis

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	v9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
)

var Redis redis

type redis struct {
	once        sync.Once
	initialized bool
	Client      *v9.Client
	Online      *v9.Client
	Cache       *v9.Client
}

func (r *redis) Setup() error {
	var setErr error
	r.once.Do(func() {
		if r.initialized {
			//setErr = fmt.Errorf("redis already initialized")
			return
		}
		var err error
		// DPI
		r.Client, err = createClient(config.Cfg.Redis.DPI)
		if err != nil {
			zap.L().Error("Failed to Create [DPI] redis Client",
				zap.Error(err),
				zap.String("Host", config.Cfg.Redis.DPI.Host),
				zap.String("Port", config.Cfg.Redis.DPI.Port),
				zap.String("Password", config.Cfg.Redis.DPI.Password),
			)
			setErr = err
			return
		}
		// Online
		r.Online, err = createClient(config.Cfg.Redis.Online)
		if err != nil {
			zap.L().Error("Failed to Create [Online] redis Client",
				zap.Error(err),
				zap.String("Host", config.Cfg.Redis.Online.Host),
				zap.String("Port", config.Cfg.Redis.Online.Port),
				zap.String("Password", config.Cfg.Redis.Online.Password),
			)
			setErr = err
			return
		}
		// Cache
		r.Cache, err = createClient(config.Cfg.Redis.Cache)
		if err != nil {
			zap.L().Error("Failed to Create [Cache] redis Client",
				zap.Error(err),
				zap.String("Host", config.Cfg.Redis.Cache.Host),
				zap.String("Port", config.Cfg.Redis.Cache.Port),
				zap.String("Password", config.Cfg.Redis.Cache.Password),
			)
			setErr = err
			return
		}
		r.initialized = true
	})
	return setErr
}

func createClient(c config.RedisConfig) (*v9.Client, error) {
	if c.Host == "" {
		return nil, fmt.Errorf("redis.host is empty")
	}
	if c.Port == "" {
		return nil, fmt.Errorf("redis.port is empty")
	}

	ctx := context.Background()
	rdb := v9.NewClient(&v9.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
	})

	// Ping redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	} else {
		zap.L().Info(
			"Connected to redis!",
			zap.String("host", c.Host),
			zap.String("port", c.Port),
			zap.String("Password", c.Password),
		)
		return rdb, nil
	}
}

func (r *redis) GetClient() *v9.Client {
	if err := r.Setup(); err != nil {
		return nil
	}
	return r.Client
}

func (r *redis) GetOnlineClient() *v9.Client {
	if err := r.Setup(); err != nil {
		return nil
	}
	return r.Online
}

func (r *redis) GetCacheClient() *v9.Client {
	if err := r.Setup(); err != nil {
		return nil
	}
	return r.Cache
}

func GetRedisClient() *v9.Client {
	return Redis.GetClient()
}

func GetOnlineRedisClient() *v9.Client {
	return Redis.GetOnlineClient()
}

func GetCacheRedisClient() *v9.Client {
	return Redis.GetCacheClient()
}
