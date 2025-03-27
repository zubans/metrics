package storage

import (
	"sort"
	"strconv"
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

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.gauges[name]
	return value, exists
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, exists := m.counters[name]
	return value, exists
}

func (m *MemStorage) GetGauges() map[string]float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]float64)
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

func (m *MemStorage) GetCounters() map[string]int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make(map[string]int64)
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

func (m *MemStorage) ShowMetrics() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var keys []string
	for k := range m.gauges {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := "<html><span style='float:left; margin-right:20px;'><strong>Guages</strong><table border=1 style='border-collapse: collapse;'><tr>"

	for _, k := range keys {
		v := m.gauges[k]
		result += "<tr><td>" + k + "</td><td>" + strconv.FormatFloat(v, 'f', 2, 64) + "</td></tr>"
	}

	result += "</table> </span>"
	result += "<span style='float:left;'><strong>Counters</strong><table border=1 style='border-collapse: collapse;'><tr>"

	for k, w := range m.counters {
		result += "<td>" + k + "</td>" + "<td>" + strconv.FormatInt(w, 10) + "</td></tr>"
	}
	result += "</table></span></html>"

	return result
}
