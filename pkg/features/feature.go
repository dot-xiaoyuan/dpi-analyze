package features

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"sync"
)

//go:embed feature.cfg
var FeatureCfg []byte

var (
	one           sync.Once
	AppFeature    []Feature
	DomainFeature []string
	DomainMap     = make(map[int]string)
	DomainAc      *ahocorasick.Matcher
)

type Feature struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Hostname string `json:"hostname"`
	Request  string `json:"request"`
	Dict     string `json:"dict"`
}

func Setup() {
	one.Do(func() {
		loadFeature()
	})
}

func loadFeature() {
	spinners.Start()
	scanner := bufio.NewScanner(bytes.NewReader(FeatureCfg))
	l := 0
	for scanner.Scan() {
		l++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		line = strings.TrimSpace(line)
		err := parse(line)
		if err != nil {
			zap.L().Error("Failed to parse feature", zap.String("line", line), zap.Error(err))
			continue
		}
	}
	spinners.Stop()
	zap.L().Info(i18n.T("Feature component initialized!"), zap.Int("count", l))
}

// 解析特征库
func parse(line string) error {
	re := regexp.MustCompile(`(\d+) (.+):\[(.+)]`)
	match := re.FindStringSubmatch(line)
	if len(match) == 0 || len(match) < 3 {
		return errors.New("invalid format Feature")
	}

	f := Feature{}
	f.ID = match[1]
	f.Name = match[2]
	temp := match[3]
	features := strings.Split(temp, ",")
	for _, item := range features {
		feature := strings.Split(item, ";")
		if len(feature) != 6 {
			continue
		}

		f.Protocol = feature[0]
		f.SrcPort = feature[1]
		f.DstPort = feature[2]
		f.Hostname = feature[3]
		f.Request = feature[4]
		f.Dict = feature[5]

		domain(f.Name, f.Hostname)

		//zap.L().Debug("append feature",
		//	zap.String("n", f.Name),
		//	zap.String("p", f.Protocol),
		//	zap.String("h", f.Hostname),
		//	zap.String("r", f.Request),
		//	zap.String("sp", f.SrcPort),
		//	zap.String("dp", f.DstPort),
		//	zap.String("d", f.Dict),
		//)
		AppFeature = append(AppFeature, f)
	}
	DomainAc = ahocorasick.NewStringMatcher(DomainFeature)
	return nil
}

func domain(app, hostname string) {
	if len(hostname) == 0 {
		return
	}
	// 加载完整域名
	DomainFeature = append(DomainFeature, hostname)
	DomainMap[len(DomainFeature)-1] = app
	// 处理多级域名
	dot := strings.Count(hostname, ".")
	if dot >= 2 {
		// 去除根域名
		parts := strings.Split(hostname, ".")
		hostname = strings.TrimSuffix(hostname, "."+parts[len(parts)-1])
		// zap.L().Info("hostname", zap.String("hostname", hostname))
		if strings.HasPrefix(hostname, ".") {
			hostname = strings.TrimPrefix(hostname, ".")
		}
		// zap.L().Info("hostname", zap.String("hostname", hostname))
		DomainFeature = append(DomainFeature, hostname)
		DomainMap[len(DomainFeature)-1] = app
	}
}
