package storage

import (
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mutex    sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.counters[name] += value
}
