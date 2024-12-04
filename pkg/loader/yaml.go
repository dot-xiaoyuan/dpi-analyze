package loader

import (
	"os"
)

type YamlLoader struct {
	Filename string
}

func (yl *YamlLoader) Load() ([]byte, error) {
	return os.ReadFile(yl.Filename)
}

func (yl *YamlLoader) Exists() bool {
	if _, err := os.Open(yl.Filename); err != nil {
		return false
	}
	return true
}
