package cache

import (
	"sync"

	"github.com/eduardoArequipa/semantic-validator/internal/metrics"
	"github.com/eduardoArequipa/semantic-validator/internal/validation"
)

type Memory struct {
	mu      sync.RWMutex
	entries map[string]validation.Result
	metrics *metrics.Metrics
}

func NewMemory() *Memory {
	return NewMemoryWithMetrics(nil)
}

func NewMemoryWithMetrics(observability *metrics.Metrics) *Memory {
	return &Memory{entries: make(map[string]validation.Result), metrics: observability}
}

func (m *Memory) Get(key string) (validation.Result, bool) {
	m.mu.RLock()
	result, found := m.entries[key]
	m.mu.RUnlock()
	if m.metrics != nil {
		if found {
			m.metrics.ObserveCacheHit()
		} else {
			m.metrics.ObserveCacheMiss()
		}
	}
	return result, found
}

func (m *Memory) Set(key string, result validation.Result) {
	m.mu.Lock()
	m.entries[key] = result
	m.mu.Unlock()
}
