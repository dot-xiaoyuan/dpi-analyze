package i18n

import (
	"embed"
	"fmt"
	v2 "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
	"sync"
)

var I18n i18n

type i18n struct {
	once        sync.Once
	initialized bool
	Lang        string
	localize    *v2.Localizer
}

//go:embed *.toml
var LocaleFS embed.FS

func (i *i18n) Setup() error {
	var setupErr error
	i.once.Do(func() {
		if i.initialized {
			setupErr = fmt.Errorf("i18n already initialized")
			return
		}
		bundle := v2.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		// 加载嵌入的翻译文件
		_, err := bundle.LoadMessageFileFS(LocaleFS, "en.toml")
		if err != nil {
			setupErr = err
			return
		}

		_, err = bundle.LoadMessageFileFS(LocaleFS, "zh-CN.toml")
		if err != nil {
			setupErr = err
			return
		}

		i.localize = v2.NewLocalizer(bundle, i.Lang)
		i.initialized = true
	})
	return setupErr
}

func (i *i18n) T(m string, templateData map[string]interface{}) string {
	localizedMsg, err := i.localize.Localize(&v2.LocalizeConfig{
		MessageID:    m,
		TemplateData: templateData,
	})
	if err != nil || localizedMsg == "" {
		return m
	}
	return localizedMsg
}

func T(m string) string {
	return I18n.T(m, nil)
}

func TT(m string, templateData map[string]interface{}) string {
	return I18n.T(m, templateData)
}
