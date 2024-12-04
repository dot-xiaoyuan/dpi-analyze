package matcher

import (
	"sync"

	"github.com/cloudflare/ahocorasick"
)

// Matcher 负责 Aho-Corasick 匹配器的管理
type Matcher struct {
	mu       sync.RWMutex
	matcher  *ahocorasick.Matcher
	patterns []string
}

// NewMatcher 创建新的 Matcher 实例
func NewMatcher(patterns []string) *Matcher {
	return &Matcher{
		matcher:  ahocorasick.NewStringMatcher(patterns),
		patterns: patterns,
	}
}

// UpdatePatterns 更新匹配器的模式
func (m *Matcher) UpdatePatterns(patterns []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matcher = ahocorasick.NewStringMatcher(patterns)
	m.patterns = patterns
}

// Match 匹配字符串，返回匹配的索引
func (m *Matcher) Match(input string) []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.matcher.Match([]byte(input))
}

// GetPatterns 返回当前匹配器的所有模式
func (m *Matcher) GetPatterns() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.patterns
}
