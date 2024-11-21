package policy

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/redis"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	Policy         *policy
	productsFields = []string{"products_id", "products_name", "mgr_name"}
	controlsFields = []string{"control_name", "disable_proxy", "proxy_times", "proxy_disable_time"}
)

func Setup() error {
	Policy = &policy{}
	return Policy.Setup()
}

// GetList 获取策略列表
func GetList() []types.Products {
	var productsList []types.Products

	for _, product := range Policy.products {
		productsList = append(productsList, product)
	}

	return productsList
}

func Get(product string) types.Products {
	return Policy.products[product]
}

type policy struct {
	once        sync.Once
	initialized bool
	products    map[string]types.Products
	controls    map[string]types.Controls
}

func (p *policy) Setup() error {
	var setupErr error
	p.once.Do(func() {
		if p.initialized {
			setupErr = fmt.Errorf("policy already initialized")
			return
		}
		err := p.load()
		if err != nil {
			setupErr = err
			return
		}
		err = p.searchMongo()
		if err != nil {
			setupErr = err
			return
		}
		zap.L().Info("加载产品和控制策略完成", zap.Int("产品数量", len(p.products)), zap.Int("控制策略", len(p.controls)))
	})
	return setupErr
}

func (p *policy) load() error {
	rdb := redis.GetUsersRedisClient()
	ctx := context.TODO()

	// 获取控制策略总数初始化map
	controlsCount := rdb.LLen(ctx, types.ListControl).Val()
	p.controls = make(map[string]types.Controls, controlsCount)

	// 遍历控制策略
	controls := rdb.LRange(ctx, types.ListControl, 0, -1).Val()
	if len(controls) == 0 {
		zap.L().Error("获取控制策略队列为空", zap.String("controls", types.ListControl))
		return errors.New("控制策略队列为空")
	}
	// 获取控制策略详情
	for _, controlID := range controls {
		var control types.Controls
		zap.L().Debug("key", zap.String("key", fmt.Sprintf(types.HashControl, controlID)), zap.Strings("fields", controlsFields))
		err := rdb.HMGet(ctx, fmt.Sprintf(types.HashControl, controlID), controlsFields...).Scan(&control)
		if err != nil {
			zap.L().Error("获取控制策略失败", zap.String("control", controlID), zap.Error(err))
			continue
		}
		p.controls[controlID] = control
	}

	// 获取产品总数初始化map
	productsCount := rdb.LLen(ctx, types.ListProducts).Val()
	p.products = make(map[string]types.Products, productsCount)

	// 遍历产品列表
	products := rdb.LRange(ctx, types.ListProducts, 0, -1).Val()
	if len(products) == 0 {
		zap.L().Error("获取产品队列为空，请检查是否配置正确", zap.String("products", types.ListProducts))
		return errors.New("获取产品队列为空，请检查是否配置正确")
	}
	// 获取产品详情
	for _, productID := range products {
		var product types.Products
		err := rdb.HMGet(ctx, fmt.Sprintf(types.HashProducts, productID), productsFields...).Scan(&product)
		if err != nil {
			zap.L().Error("获取产品信息失败", zap.String("product", productID), zap.Error(err))
			continue
		}
		// 获取产品绑定的控制策略
		productControl := rdb.LIndex(ctx, fmt.Sprintf(types.ListProductsControl, productID), 0).Val()
		product.Controls = p.controls[productControl]
		p.products[productID] = product
	}
	return nil
}

func (p *policy) searchMongo() error {
	// 检查集合是否有数据
	count, err := mongo.GetMongoClient().
		Database(types.MongoDatabaseConfigs).
		Collection(types.MongoCollectionPolicy).
		CountDocuments(mongo.Context, bson.M{})
	if err != nil {
		zap.L().Error("Failed to count documents in MongoDB", zap.Error(err))
		return err
	}

	if count == 0 {
		zap.L().Info("MongoDB is empty, storing configurations from Redis")
		return p.storeMongo() // 持久化 Redis 数据
	}

	// 获取已有的配置
	cursor, err := mongo.GetMongoClient().
		Database(types.MongoDatabaseConfigs).
		Collection(types.MongoCollectionPolicy).
		Find(mongo.Context, bson.M{})
	if err != nil {
		zap.L().Error("Failed to find documents in MongoDB", zap.Error(err))
		return err
	}

	var configs []types.Products
	err = cursor.All(mongo.Context, &configs)
	if err != nil {
		zap.L().Error("Failed to decode MongoDB documents", zap.Error(err))
		return err
	}

	// 将 MongoDB 的配置更新到内存
	for _, config := range configs {
		product, exists := p.products[config.ProductsID]
		if exists {
			// 更新策略到内存中的产品
			product.Policy = config.Policy
			p.products[config.ProductsID] = product
		} else {
			// 如果 Redis 没有该产品，直接从 MongoDB 添加
			p.products[config.ProductsID] = config
		}
	}

	zap.L().Info("Loaded policies from MongoDB",
		zap.Int("products_count", len(p.products)),
		zap.Int("controls_count", len(p.controls)))
	return nil
}

func (p *policy) storeMongo() error {
	client := mongo.GetMongoClient()
	collection := client.Database(types.MongoDatabaseConfigs).Collection(types.MongoCollectionPolicy)

	for _, product := range p.products {
		if product.ALL == 0 {
			product.Policy.ALL = 4
		}
		if product.Mobile == 0 {
			product.Policy.Mobile = 2
		}
		if product.Pc == 0 {
			product.Policy.Pc = 2
		}
		p.products[product.ProductsID] = product
		config := product

		filter := bson.M{"_id": product.ProductsID}
		update := bson.M{
			"$set":         config,
			"$setOnInsert": bson.M{"created_at": time.Now()},
		}

		_, err := collection.UpdateOne(mongo.Context, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			zap.L().Error("Failed to upsert product",
				zap.String("product_id", product.ProductsID),
				zap.Error(err))
			return err
		}
	}
	zap.L().Info("Successfully stored policies in MongoDB")
	return nil
}

func (p *policy) Update(params types.Products) error {
	product, ok := p.products[params.ProductsID]
	if !ok {
		return errors.New("product not found")
	}
	product.Pc = params.Pc
	product.Mobile = params.Mobile
	product.ALL = params.Pc + params.Mobile
	p.products[params.ProductsID] = product
	zap.L().Debug("product updated", zap.Any("product", p.products[params.ProductsID]))
	return p.storeMongo()
}
