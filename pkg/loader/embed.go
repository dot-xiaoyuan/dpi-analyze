package loader

import "embed"

type EmbedLoader struct {
	Fs       embed.FS
	Filename string
}

func (el *EmbedLoader) Load() ([]byte, error) {
	return el.Fs.ReadFile(el.Filename)
}

func (el *EmbedLoader) Exists() bool {
	if _, err := el.Fs.Open(el.Filename); err != nil {
		return false
	}
	return true
}
