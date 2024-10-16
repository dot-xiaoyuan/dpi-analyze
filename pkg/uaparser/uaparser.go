package uaparser

import (
	_ "embed"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
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
	defer func() {
		if err := recover(); err != nil {
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
}

func Parse(ua string) string {
	if Parser == nil {
		Setup()
	}
	client := Parser.ParseOs(ua)
	return client.Family
}
