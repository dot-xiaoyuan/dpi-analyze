package uaparser

import (
	_ "embed"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
	"sync"
)

//go:embed regexes.yaml
var regexes []byte

var (
	one    sync.Once
	Parser *uaparser.Parser
)

func Setup() error {
	var setupErr error
	one.Do(func() {
		err := loadParser()
		setupErr = err
	})
	return setupErr
}

func loadParser() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			zap.L().Error(i18n.T("Failed to load ua parser"), zap.Any("error", err))
			return
		}
	}()

	Parser, err = uaparser.NewFromBytes(regexes)
	if err != nil {
		return err
	}
	return nil
}

func Parse(ua string) string {
	if Parser == nil {
		Setup()
	}
	client := Parser.ParseOs(ua)
	return client.Family
}
