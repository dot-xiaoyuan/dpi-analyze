package loader

import (
	"errors"
	"fmt"
)

type Manager struct {
	Embed *EmbedLoader
	Mongo *MongoLoader
	Yaml  *YamlLoader
}

// Load 加载
func (m *Manager) Load() ([]byte, error) {
	// 优先加载mongo，没有再加载embed
	if m.Mongo != nil && m.Mongo.Exists() {
		return m.Mongo.Load()
	}
	if m.Yaml != nil && m.Yaml.Exists() {
		data, err := m.Yaml.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load from yaml file: %w", err)
		}

		// 将加载的数据存储到 MongoDB
		if m.Mongo != nil {
			err = m.Mongo.Save(data, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to store data into MongoDB: %w", err)
			}
		}

		return data, nil
	}
	if m.Embed != nil && m.Embed.Exists() {
		data, err := m.Embed.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load from embedded file: %w", err)
		}

		// 将加载的数据存储到 MongoDB
		if m.Mongo != nil {
			err = m.Mongo.Save(data, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to store data into MongoDB: %w", err)
			}
		}

		return data, nil
	}
	return nil, errors.New("no data source available")
}

// Version 版本
func (m *Manager) Version() string {
	v, _ := m.Mongo.GetCurrentVersion()
	return v
}

// History 历史版本
func (m *Manager) History() []History {
	v, _ := m.Mongo.GetHistoryVersions()
	return v
}
