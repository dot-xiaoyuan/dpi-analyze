package maxmind

import (
	_ "embed"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
	"net"
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
	Geo2IPDB, err = maxminddb.Open(file)
	if err != nil {
		panic(err)
	}
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
