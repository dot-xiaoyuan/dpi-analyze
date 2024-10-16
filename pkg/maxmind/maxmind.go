package maxmind

import (
	_ "embed"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
	"net"
	"os"
	"sync"
)

var (
	one      sync.Once
	Geo2IPDB *maxminddb.Reader
)

func Setup(file string) error {
	one.Do(func() {
		NewGeoIP(file)
	})
	return nil
}

func NewGeoIP(file string) {
	var err error
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error(i18n.T("Failed to load GeoIP database."), zap.Any("error", err))
			os.Exit(1)
		}
	}()
	Geo2IPDB, err = maxminddb.Open(file)
	if err != nil {
		panic(err)
	}
	zap.L().Info(i18n.T("Geo2IP component initialized!"))
}

type Record struct {
	City struct {
		GeoNameID uint32            `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
}

func GetCity(ip net.IP) string {
	if Geo2IPDB == nil {
		zap.L().Debug("geo2ip db is nil")
	}
	defer Geo2IPDB.Close()

	var r Record
	err := Geo2IPDB.Lookup(ip, &r)
	if err != nil {
		return ""
	}
	if v, ok := r.City.Names["zh-CN"]; ok {
		return v
	}
	return r.City.Names["en"]
}
