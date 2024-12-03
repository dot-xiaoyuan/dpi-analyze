package loader

import (
	"errors"
	"fmt"
)

type Manager struct {
	Embed *EmbedLoader
	Mongo *MongoLoader
}

func (m *Manager) Load() ([]byte, error) {
	// 优先加载mongo，没有再加载embed
	if m.Mongo.Exists() {
		return m.Mongo.Load()
	}
	if m.Embed.Exists() {
		data, err := m.Embed.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load from embedded file: %w", err)
		}

		// 将加载的数据存储到 MongoDB
		if m.Mongo != nil {
			err = m.Mongo.Save(data)
			if err != nil {
				return nil, fmt.Errorf("failed to store data into MongoDB: %w", err)
			}
		}

		return data, nil
	}
	return nil, errors.New("no data source available")
}
