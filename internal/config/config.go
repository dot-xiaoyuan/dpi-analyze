package config

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"os"
	"strings"
)

var (
	Translate *i18n.Translator
	Language  string
)

func Setup() {
	// 初始化翻译
	Translate = i18n.NewTranslator(getSystemLanguage())
}

// 获取系统语言
func getSystemLanguage() string {
	if Language != "" && (Language == "en" || Language == "zh-CN") {
		return Language
	}
	lang := os.Getenv("LANG")
	if lang != "" {
		return strings.Split(lang, ".")[0]
	}
	return "en"
}
