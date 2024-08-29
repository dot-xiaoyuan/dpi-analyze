package stream

import "sync"

type Reader interface {
	Run(wg *sync.WaitGroup)
}
