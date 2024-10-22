package maxmind

import (
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
	"net"
	"sync"
)

var MaxMind maxmind

type maxmind struct {
	once        sync.Once
	initialized bool
	Filename    string
	Geo2IP      *maxminddb.Reader
}

func (m *maxmind) Setup() error {
	var setupErr error
	m.once.Do(func() {
		if m.initialized {
			setupErr = fmt.Errorf("maxmind already initialized")
			return
		}
		var err error
		m.Geo2IP, err = maxminddb.Open(m.Filename)
		if err != nil {
			zap.L().Error("Failed to open GeoIP database", zap.String("filename", m.Filename), zap.Error(err))
			setupErr = err
			return
		}
		m.initialized = true
	})
	return setupErr
}

type Record struct {
	City struct {
		GeoNameID uint32            `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
}

func (m *maxmind) GetCity(ip net.IP) string {
	if m.Geo2IP == nil {
		if err := m.Setup(); err != nil {
			zap.L().Error("Failed to setup GeoIP database", zap.String("filename", m.Filename), zap.Error(err))
			return ""
		}
	}
	defer m.Geo2IP.Close()

	var r Record
	err := m.Geo2IP.Lookup(ip, &r)
	if err != nil {
		return ""
	}
	if v, ok := r.City.Names["zh-CN"]; ok {
		return v
	}
	return r.City.Names["en"]
}
