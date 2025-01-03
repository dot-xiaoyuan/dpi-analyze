package application

import (
	"embed"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/loader"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/matcher"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/parser"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/statictics"
	"go.uber.org/zap"
	"os"
	"strings"
	"sync"
)

var (
	//go:embed feature2.0_cn_24.10.14-plus.cfg
	feature         embed.FS
	LoaderManger    = loader.Manager{}
	Feature         []string
	AppMap          = make(map[int]parser.Application)
	MatcherInstance *matcher.Matcher
	mutex           sync.Mutex
)

func Setup() error {
	LoaderManger.Embed = &loader.EmbedLoader{
		Fs:       feature,
		Filename: "feature2.0_cn_24.10.14-plus.cfg",
	}
	LoaderManger.Mongo = &loader.MongoLoader{
		Client:             mongodb.GetMongoClient(),
		MetadataCollection: types.MongoCollectionFeatureApplication,
		HistoryCollection:  types.MongoCollectionFeatureApplicationHistory,
		Database:           types.MongoDatabaseConfigs,
	}
	data, err := LoaderManger.Load()
	if err != nil {
		return err
	}
	// 调用解析逻辑
	err = Parse(data)
	if err != nil {
		return err
	}

	initMatcher()
	return nil
}

// 初始化 Aho-Corasick 匹配器
func initMatcher() {
	mutex.Lock()
	defer mutex.Unlock()
	MatcherInstance = matcher.NewMatcher(Feature)
	zap.L().Info("Initialized Aho-Corasick matcher", zap.Int("patternCount", len(Feature)))
}

// Parse 解析
func Parse(data []byte) error {
	applications, err := parser.ParseApplications(data)
	if err != nil {
		zap.L().Error("Failed to parse domains", zap.Error(err))
		return err
	}

	// 保存解析结果
	mutex.Lock()
	defer mutex.Unlock()
	for _, app := range applications {
		domainParse(app)
	}

	return nil
}

// Match 匹配
func Match(input string) (ok bool, app parser.Application) {
	hits := MatcherInstance.Match(input)
	if hits == nil {
		return false, parser.Application{}
	}
	if app, ok = AppMap[hits[0]]; ok {
		statictics.Application.Increment(app.Name)
		statictics.AppCategory.Increment(app.Category)
		return true, app
	}
	return false, parser.Application{}
}

// Update 更新
func Update(filepath string) error {
	file, err := os.ReadFile(filepath)
	if err != nil {
		zap.L().Error("Failed to open domain file", zap.String("file", filepath), zap.Error(err))
		return err
	}

	data, err := parser.ParseApplications(file)
	if err != nil {
		zap.L().Error("Failed to parse domain file", zap.String("file", filepath), zap.Error(err))
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, app := range data {
		domainParse(app)
	}
	err = LoaderManger.Mongo.Save(file, len(Feature)-len(data))
	if err != nil {
		zap.L().Error("Failed to save domain file", zap.String("file", filepath), zap.Error(err))
		return err
	}
	return nil
}

// 域名解析
func domainParse(a parser.Application) {
	if a.Hostname == "" {
		return
	}

	Feature = append(Feature, a.Hostname)
	AppMap[len(Feature)-1] = a
	// 处理二级域名
	if strings.Count(a.Hostname, ".") >= 2 {
		parts := strings.SplitN(a.Hostname, ".", 2)
		subdomain := parts[1]
		if ignoreDomain(subdomain) {
			return
		}
		Feature = append(Feature, subdomain)
		AppMap[len(Feature)-1] = a
	}
}

// 忽略域名
func ignoreDomain(domain string) bool {
	for _, item := range config.Cfg.IgnoreFeature {
		if strings.Contains(domain, item) {
			return true
		}
	}
	return false
}
