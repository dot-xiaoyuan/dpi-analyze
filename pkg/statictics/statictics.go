package statictics

import (
	"sync"
	"sync/atomic"
)

// 统计

var (
	Application      Statics
	AppCategory      Statics
	ApplicationLayer Statics
	TransportLayer   Statics
)

type Statics struct {
	sync.Map
}

type Charts struct {
	Name  string `json:"type"`
	Value int64  `json:"value"`
}

func (s *Statics) Increment(key string) {
	actual, _ := s.LoadOrStore(key, new(int64))
	atomic.AddInt64(actual.(*int64), 1)
}

func (s *Statics) GetStats() []Charts {
	var result []Charts
	index := 0
	s.Range(func(key, value any) bool {
		if index == 50 {
			return true
		}
		result = append(result, Charts{
			Name:  key.(string),
			Value: atomic.LoadInt64(value.(*int64)),
		})
		index++
		return true
	})
	return result
}
