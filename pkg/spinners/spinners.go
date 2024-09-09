package spinners

import (
	"github.com/briandowns/spinner"
	"os"
	"sync"
	"time"
)

var (
	one     sync.Once
	Spinner *spinner.Spinner
)

func Setup() {
	one.Do(func() {
		loadSpinner()
	})
}

func loadSpinner() {
	Spinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
}

func Start() {
	Spinner.Start()
}

func Stop() {
	Spinner.Stop()
}
