package uaparser

import (
	_ "embed"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/spinners"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
	"os"
	"sync"
)

//go:embed regexes.yaml
var regexes []byte

var (
	one    sync.Once
	Parser *uaparser.Parser
)

func Setup() {
	one.Do(func() {
		loadParser()
	})
}

func loadParser() {
	spinners.Start()

	defer func() {
		if err := recover(); err != nil {
			spinners.Stop()
			zap.L().Error(i18n.T("Failed to load ua parser"), zap.Any("error", err))
			os.Exit(1)
		}
	}()

	var err error
	Parser, err = uaparser.NewFromBytes(regexes)
	if err != nil {
		panic(err)
	}
	zap.L().Info(i18n.T("ua parser component initialized!"))
	spinners.Stop()
	//uagent := "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_3; en-us; Silk/1.1.0-80) AppleWebKit/533.16 (KHTML, like Gecko) Version/5.0 Safari/533.16 Silk-Accelerated=true"
	//client := Parser.Parse(uagent)
	//zap.L().Debug("ua parser res", zap.Any("c", client))
}

func Parse(ua string) string {
	client := Parser.ParseOs(ua)
	return client.Family
}
