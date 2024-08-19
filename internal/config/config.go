package config

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	Translate *i18n.Translator
	Cmd       = &cobra.Command{}
)

func init() {
	// 初始化翻译
	Translate = i18n.NewTranslator(getSystemLanguage())

	Cmd = &cobra.Command{
		Use:     "config",
		Short:   Translate.T("Configure related commands", nil),
		Version: "1.0.0",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func setup() {

}

// 获取系统语言
func getSystemLanguage() string {
	lang := os.Getenv("LANG")
	if lang != "" {
		return strings.Split(lang, ".")[0]
	}
	return "en"
}
