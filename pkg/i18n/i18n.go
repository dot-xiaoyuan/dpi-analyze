package i18n

import (
	"embed"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
	"sync"
)

//go:embed *.toml
var LocaleFS embed.FS

var (
	Translate *Translator
	one       sync.Once
)

// Translator 翻译管理结构体
type Translator struct {
	loc *i18n.Localizer
}

func Setup(l string) {
	one.Do(func() {
		Translate = NewTranslator(l)
	})
}

// NewTranslator 初始化方法
func NewTranslator(lang string) *Translator {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 加载嵌入的翻译文件
	_, err := bundle.LoadMessageFileFS(LocaleFS, "en.toml")
	if err != nil {
		panic(err)
	}

	_, err = bundle.LoadMessageFileFS(LocaleFS, "zh-CN.toml")
	if err != nil {
		panic(err)
	}

	loc := i18n.NewLocalizer(bundle, lang)
	return &Translator{loc: loc}
}

// T 翻译方法
func (t *Translator) T(messageID string, templateData map[string]interface{}) string {
	localizedMessage, err := t.loc.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil || localizedMessage == "" {
		return messageID
	}
	return localizedMessage
}
