package features

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/cloudflare/ahocorasick"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
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
	parseMu       sync.Mutex // 用于避免并发问题
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

// Setup 初始化特征组件，确保只加载一次
func Setup() error {
	var setupErr error
	one.Do(func() {
		setupErr = loadFeature()
	})
	return setupErr
}

// 加载特征配置文件并解析
func loadFeature() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			zap.L().Error("Recovered from panic in loadFeature", zap.Any("error", r))
			return
		}
	}()
	scanner := bufio.NewScanner(bytes.NewReader(FeatureCfg))
	lineNumber := 0
	var wg sync.WaitGroup

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNumber++

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		wg.Add(1)
		go func(line string, lineNum int) {
			defer wg.Done()
			if err = parse(line); err != nil {
				zap.L().Error("Failed to parse feature", zap.String("line", line), zap.Error(err), zap.Int("lineNumber", lineNum))
			}
		}(line, lineNumber)
	}

	wg.Wait()
	if scanner.Err() != nil {
		return fmt.Errorf("error reading feature file: %w", scanner.Err())
	}

	// 创建 Aho-Corasick 匹配器
	DomainAc = ahocorasick.NewStringMatcher(DomainFeature)
	zap.L().Info(i18n.T("Feature component initialized!"), zap.Int("count", len(AppFeature)))

	return nil
}

// 解析单行特征配置
func parse(line string) error {
	re := regexp.MustCompile(`(\d+) (.+):\[(.+)]`)
	match := re.FindStringSubmatch(line)
	if len(match) < 4 {
		return errors.New("invalid feature format")
	}

	f := Feature{
		ID:   match[1],
		Name: match[2],
	}

	features := strings.Split(match[3], ",")
	for _, item := range features {
		feature := strings.Split(item, ";")
		if len(feature) != 6 {
			return errors.New("invalid feature details")
		}

		f.Protocol, f.SrcPort, f.DstPort = feature[0], feature[1], feature[2]
		f.Hostname, f.Request, f.Dict = feature[3], feature[4], feature[5]

		parseMu.Lock()
		AppFeature = append(AppFeature, f)
		addDomain(f.Name, f.Hostname)
		parseMu.Unlock()
	}

	return nil
}

// 处理域名并添加到匹配器列表
func addDomain(app, hostname string) {
	if hostname == "" {
		return
	}

	DomainFeature = append(DomainFeature, hostname)
	DomainMap[len(DomainFeature)-1] = app

	// 处理多级域名
	if strings.Count(hostname, ".") >= 2 {
		parts := strings.SplitN(hostname, ".", 2)
		subdomain := parts[1]
		DomainFeature = append(DomainFeature, subdomain)
		DomainMap[len(DomainFeature)-1] = app
	}
}
