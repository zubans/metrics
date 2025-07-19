package storage

import (
	"context"
	"sync"

	"github.com/zubans/metrics/internal/models"
)

// MemStorage реализует хранение метрик в памяти с потокобезопасностью.
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
	mutex    sync.Mutex
}

// NewMemStorage создаёт новое in-memory хранилище метрик.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

// UpdateGauge обновляет gauge-метрику.
func (m *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Gauges[name] = value

	return m.Gauges[name]
}

// UpdateCounter обновляет counter-метрику.
func (m *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Counters[name] += value

	return m.Counters[name]
}

// GetGauge возвращает значение gauge-метрики.
func (m *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Gauges[name]
	return value, exists
}

// GetCounter возвращает значение counter-метрики.
func (m *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.Counters[name]
	return value, exists
}

// GetGauges возвращает копию всех gauge-метрик.
func (m *MemStorage) GetGauges(ctx context.Context) map[string]float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]float64)
	for k, v := range m.Gauges {
		result[k] = v
	}
	return result
}

// GetCounters возвращает копию всех counter-метрик.
func (m *MemStorage) GetCounters(ctx context.Context) map[string]int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]int64)
	for k, v := range m.Counters {
		result[k] = v
	}
	return result
}

// ShowMetrics возвращает все метрики.
func (m *MemStorage) ShowMetrics(ctx context.Context) (map[string]float64, map[string]int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Gauges, m.Counters, nil
}

func (m *MemStorage) UpdateMetrics(ctx context.Context, mDTO []models.MetricsDTO) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, v := range mDTO {
		switch v.MType {
		case string(models.Counter):
			if v.Delta != nil {
				m.Counters[v.ID] += *v.Delta
			}
		case string(models.Gauge):
			if v.Value != nil {
				m.Gauges[v.ID] += *v.Value
			}
		}
	}
	return nil
}
