package features

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"sync"
)

//go:embed feature.cfg
var FeatureCfg []byte

var (
	Features      features
	AhoCorasick   *ahocorasick.Matcher
	AppFeature    []Feature               // 应用特征切片
	DomainFeature []string                // 域名特征切片
	DomainMap     = make(map[int]Feature) // 域名
	parseMutex    sync.Mutex
)

func Setup() error {
	return Features.Setup()
}

type Feature struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Hostname string `json:"hostname"`
	Request  string `json:"request"`
	Dict     string `json:"dict"`
	Category string `json:"category"`
}

type features struct {
	once        sync.Once
	initialized bool
}

func (f *features) Setup() error {
	var setupErr error
	f.once.Do(func() {
		if f.initialized {
			setupErr = fmt.Errorf("features already initialized")
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(FeatureCfg))
		lineNumber := 0
		var wg sync.WaitGroup
		var category string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			lineNumber++

			// 跳过空行
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				if strings.Contains(line, "class") {
					temp := strings.Split(line, " ")
					category = temp[len(temp)-1]
					continue
				} else {
					continue
				}
			}

			wg.Add(1)
			var err error
			go func(line string, lineNum int, category string) {
				defer wg.Done()
				if err = parse(line, category); err != nil {
					zap.L().Error("Failed to parse feature", zap.String("line", line), zap.Error(err), zap.Int("lineNumber", lineNum))
				}
			}(line, lineNumber, category)
		}

		wg.Wait()
		if scanner.Err() != nil {
			setupErr = fmt.Errorf("error reading feature file: %w", scanner.Err())
			return
		}

		// 创建 Aho-Corasick 匹配器
		AhoCorasick = ahocorasick.NewStringMatcher(DomainFeature)
		f.initialized = true
	})
	return setupErr
}

func Match(s string) (ok bool, feature Feature) {
	hits := AhoCorasick.MatchThreadSafe([]byte(s))
	if hits == nil {
		return false, Feature{}
	}
	if feature, ok = DomainMap[hits[0]]; ok {
		//zap.L().Info("匹配到域名信息", zap.String("hostname", h), zap.String("name", name), zap.String("domain", DomainMap[hits[0]]))
		return true, feature
	}
	return false, Feature{}
}

// 解析单行特征配置
func parse(line, category string) error {
	re := regexp.MustCompile(`(\d+) (.+):\[(.+)]`)
	match := re.FindStringSubmatch(line)
	if len(match) < 4 {
		return errors.New("invalid feature format")
	}

	f := Feature{
		ID:       match[1],
		Name:     match[2],
		Category: category,
	}

	featureList := strings.Split(match[3], ",")
	for _, item := range featureList {
		feature := strings.Split(item, ";")
		if len(feature) != 6 {
			continue
		}

		f.Protocol, f.SrcPort, f.DstPort = feature[0], feature[1], feature[2]
		f.Hostname, f.Request, f.Dict = feature[3], feature[4], feature[5]

		parseMutex.Lock()
		AppFeature = append(AppFeature, f)
		addDomain(f)
		parseMutex.Unlock()
	}

	return nil
}

// 处理域名并添加到匹配器列表
func addDomain(f Feature) {
	if f.Hostname == "" {
		return
	}

	DomainFeature = append(DomainFeature, f.Hostname)
	DomainMap[len(DomainFeature)-1] = f
	// 处理多级域名
	if strings.Count(f.Hostname, ".") >= 2 {
		parts := strings.SplitN(f.Hostname, ".", 2)
		subdomain := parts[1]
		if ignoreDomain(subdomain) {
			return
		}
		DomainFeature = append(DomainFeature, subdomain)
		DomainMap[len(DomainFeature)-1] = f
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
