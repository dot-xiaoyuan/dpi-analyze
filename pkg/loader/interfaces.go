package loader

type loader interface {
	Load() ([]byte, error)
	Exists() bool
}
