package storage

import (
	"context"
	"fmt"
	"github.com/zubans/metrics/internal/models"
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

func (m *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Gauges[name] = value

	return m.Gauges[name]
}

func (m *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Counters[name] += value

	return m.Counters[name]
}

func (m *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Gauges[name]
	return value, exists
}

func (m *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Counters[name]
	return value, exists
}

func (m *MemStorage) GetGauges(ctx context.Context) map[string]float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]float64)
	for k, v := range m.Gauges {
		result[k] = v
	}
	return result
}

func (m *MemStorage) GetCounters(ctx context.Context) map[string]int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]int64)
	for k, v := range m.Counters {
		result[k] = v
	}
	return result
}

func (m *MemStorage) ShowMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Gauges, m.Counters
}

func (m *MemStorage) UpdateMetrics(ctx context.Context, mDTO []models.MetricsDTO) error {
	return fmt.Errorf("forbidden")
}
