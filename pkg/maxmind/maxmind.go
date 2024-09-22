package maxmind

import (
	_ "embed"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
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

func Setup(file string) {
	one.Do(func() {
		NewGeoIP(file)
	})
}

func NewGeoIP(file string) {
	var err error
	spinners.Start()
	defer func() {
		if err := recover(); err != nil {
			spinners.Start()
			zap.L().Error(i18n.T("Failed to load GeoIP database."), zap.Any("error", err))
			os.Exit(1)
		}
	}()
	Geo2IPDB, err = maxminddb.Open(file)
	if err != nil {
		panic(err)
	}
	zap.L().Info(i18n.T(""))
	spinners.Start()
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
