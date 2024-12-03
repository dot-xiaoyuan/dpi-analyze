package loader

import "embed"

type EmbedLoader struct {
	Fs       embed.FS
	FileName string
}

func (el *EmbedLoader) Load() ([]byte, error) {
	return el.Fs.ReadFile(el.FileName)
}

func (el *EmbedLoader) Exists() bool {
	if _, err := el.Fs.Open(el.FileName); err != nil {
		return false
	}
	return true
}
