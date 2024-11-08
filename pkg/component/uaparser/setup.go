package uaparser

import (
	_ "embed"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
	"regexp"
	"sync"
)

//go:embed regexes.yaml
var regexes []byte

var UaParser uaParser

func Setup() error {
	return UaParser.Setup()
}

type uaParser struct {
	once        sync.Once
	initialized bool
	Parser      *uaparser.Parser
}

func (u *uaParser) Setup() error {
	var setupErr error
	u.once.Do(func() {
		if u.initialized {
			setupErr = fmt.Errorf("uaparser already initialized")
			return
		}
		var err error
		u.Parser, err = uaparser.NewFromBytes(regexes)
		if err != nil {
			zap.L().Error("Failed to initialize ua parser", zap.Error(err))
			setupErr = err
			return
		}
		u.initialized = true
	})
	return setupErr
}

func (u *uaParser) Parse(ua string) (*uaparser.Client, error) {
	if u.Parser == nil {
		if err := u.Setup(); err != nil {
			return nil, err
		}
	}
	client := u.Parser.Parse(ua)
	return client, nil

}

func Parse(ua string) string {
	client, _ := UaParser.Parse(ua)
	if client.Os.ToString() == "Other" || client.Os.ToVersionString() == "" {
		return ""
	}
	zap.L().Debug(i18n.T("Parsing results"), zap.String("os", client.Os.Family), zap.String("version", client.Os.ToVersionString()), zap.String("brand", client.Device.Brand), zap.String("model", client.Device.Model))
	return fmt.Sprintf("%s-%s", client.Os.Family, client.Os.ToVersionString())
}

// Filter 过滤ua
func Filter(host string) bool {
	// 要过滤的域名的正则表达式
	filterPattern := `(?:ajax\.googleapis\.com|ajax\.microsoft\.com|cdnjs\.cloudflare\.com|code\.jquery\.com|google-analytics\.com|analytics\.google\.com|doubleclick\.net|googlesyndication\.com|ads\.linkedin\.com|facebook\.com|fbcdn\.net|connect\.facebook\.net|twitter\.com|t\.co|login\.live\.com|accounts\.google\.com|chrome\.google\.com|crashlytics\.com|safebrowsing\.googleapis\.com)`

	// 编译正则表达式
	filterRegex, err := regexp.Compile(filterPattern)
	if err != nil {
		fmt.Println("Failed to compile regex:", err)
		return false
	}
	return !filterRegex.MatchString(host)
}

func DeviceMatch(os string) {

}
