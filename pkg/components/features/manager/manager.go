package manager

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/loader"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/matcher"
	"go.uber.org/zap"
	"os"
)

// Config 配置结构
type Config struct {
	Filename              string // 配置文件
	CollectionName        string // Mongo 集合名
	HistoryCollectionName string
	DatabaseName          string                                                   // 数据库名
	ParserFunc            func(data []byte) ([]string, map[int]interface{}, error) // 数据解析函数
}

// Manager 通用管理器
type Manager struct {
	Loader          *loader.Manager     // 数据加载器
	MatcherInstance *matcher.Matcher    // Aho-Corasick 匹配器
	Feature         []string            // 匹配器模式
	Map             map[int]interface{} // 数据映射
	Config          Config              // 配置
}

// NewManager 创建通用管理器实例
func NewManager(config Config) *Manager {
	return &Manager{
		Loader: &loader.Manager{
			Mongo: &loader.MongoLoader{
				Client:             mongo.GetMongoClient(),
				MetadataCollection: config.CollectionName,
				HistoryCollection:  config.HistoryCollectionName,
				Database:           config.DatabaseName,
			},
			Yaml: &loader.YamlLoader{
				Filename: config.Filename,
			},
		},
		Config: config,
	}
}

// Setup 初始化加载器、解析数据并构建匹配器
func (m *Manager) Setup() error {
	data, err := m.Loader.Load()
	if err != nil {
		zap.L().Error("Error loading features", zap.Error(err))
		return err
	}

	// 解析数据
	features, mapping, err := m.Config.ParserFunc(data)
	if err != nil {
		zap.L().Error("Failed to parse data", zap.Error(err))
		return err
	}

	m.Feature = features
	m.Map = mapping
	m.MatcherInstance = matcher.NewMatcher(features)
	zap.L().Info("Matcher initialized", zap.String("module", m.Config.Filename), zap.Int("patternCount", len(features)))
	return nil
}

// Match 匹配输入字符串
func (m *Manager) Match(input string) (ok bool, result interface{}) {
	hits := m.MatcherInstance.Match(input)
	if len(hits) == 0 {
		return false, nil
	}

	if result, ok = m.Map[hits[0]]; ok {
		return true, result
	}
	return false, nil
}

// Update 更新
func (m *Manager) Update(filepath string) error {
	file, err := os.ReadFile(filepath)
	if err != nil {
		zap.L().Error("Failed to open domain file", zap.String("file", filepath), zap.Error(err))
		return err
	}

	features, mapping, err := m.Config.ParserFunc(file)
	if err != nil {
		zap.L().Error("Failed to parse data", zap.String("file", filepath), zap.Error(err))
		return err
	}

	err = m.Loader.Mongo.Save(file, len(m.Feature)-len(features))
	if err != nil {
		zap.L().Error("Failed to update domain file", zap.String("file", filepath), zap.Error(err))
		return err
	}
	m.Feature = features
	m.Map = mapping
	m.MatcherInstance = matcher.NewMatcher(features)
	zap.L().Info("Matcher initialized", zap.String("module", m.Config.Filename), zap.Int("patternCount", len(features)))
	return nil
}
