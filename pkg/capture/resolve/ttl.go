package resolve

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/types"
	"strconv"
	"time"
)

func AnalyzeByTTL(ip string, ttl uint8) {
	dr := types.DeviceRecord{
		IP:           ip,
		OriginChanel: types.TTL,
		OriginValue:  strconv.Itoa(int(ttl)),
		Os:           "windows",
		Version:      "",
		Device:       "windows",
		Brand:        "",
		Model:        "",
		Icon:         "icon-windows",
		Description:  "TTL 识别",
		LastSeen:     time.Now(),
	}

	Handle(dr)
}
