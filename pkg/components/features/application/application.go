package application

import (
	"embed"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/loader"
	"go.uber.org/zap"
)

var (
	//go:embed feature2.0_cn_24.10.14-plus.cfg
	feature      embed.FS
	loaderManger = loader.Manager{}
)

func Setup() error {
	loaderManger.Embed = &loader.EmbedLoader{
		Fs:       feature,
		FileName: "feature2.0_cn_24.10.14-plus.cfg",
	}
	loaderManger.Mongo = &loader.MongoLoader{
		Client:     mongodb.GetMongoClient(),
		Collection: types.MongoCollectionFeatureApplication,
		Database:   types.MongoDatabaseConfigs,
	}
	data, err := loaderManger.Load()
	if err != nil {
		return err
	}
	zap.L().Info("data length", zap.Int("length", len(data)))
	zap.L().Info("Load Feature Application success", zap.String("success", "true"), zap.Error(err))
	return err
}
