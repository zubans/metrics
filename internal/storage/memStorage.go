package storage

import (
	"sync"
)

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
	mutex    sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Counters[name] += value
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Gauges[name]
	return value, exists
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Counters[name]
	return value, exists
}

func (m *MemStorage) GetGauges() map[string]float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]float64)
	for k, v := range m.Gauges {
		result[k] = v
	}
	return result
}

func (m *MemStorage) GetCounters() map[string]int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]int64)
	for k, v := range m.Counters {
		result[k] = v
	}
	return result
}

func (m *MemStorage) ShowMetrics() (map[string]float64, map[string]int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Gauges, m.Counters
}
