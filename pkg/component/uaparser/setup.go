package uaparser

import (
	_ "embed"
	"fmt"
	"github.com/ua-parser/uap-go/uaparser"
	"go.uber.org/zap"
	"sync"
)

//go:embed regexes.yaml
var regexes []byte

var UaParser uaParser

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

func (u *uaParser) Parse(ua string) (*uaparser.Os, error) {
	if u.Parser == nil {
		if err := u.Setup(); err != nil {
			return nil, err
		}
	}
	os := u.Parser.ParseOs(ua)
	return os, nil

}

func Parse(ua string) string {
	os, _ := UaParser.Parse(ua)
	return os.Family
}
