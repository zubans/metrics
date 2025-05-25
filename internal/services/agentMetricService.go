package services

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type MetricsCollector interface {
	CollectMetrics()
	GetMetrics() *models.Metrics
}

type MetricsService struct {
	metrics        *models.Metrics
	Cfg            *config.AgentConfig
	cpuUtilization float64
	mu             sync.Mutex
}

func NewMetricsService(cfg *config.AgentConfig) *MetricsService {
	ms := &MetricsService{
		metrics: &models.Metrics{},
		Cfg:     cfg,
	}
	go ms.startCPUMonitoring()

	return ms
}

func (ms *MetricsService) startCPUMonitoring() {
	for {
		percent, err := cpu.Percent(time.Second, false)
		if err == nil && len(percent) > 0 {
			ms.mu.Lock()
			ms.cpuUtilization = percent[0]
			ms.mu.Unlock()
		}
		time.Sleep(time.Second)
	}
}

func (ms *MetricsService) CollectMetrics() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ms.metrics.PollCount++

	ms.addMetrics(memStats)

	if v, err := mem.VirtualMemory(); err == nil {
		total, free := int64(v.Total), int64(v.Free)
		ms.metrics.MetricList = append(ms.metrics.MetricList,
			models.Metric{Name: "TotalMemory", Type: "gauge", Value: int(total)},
			models.Metric{Name: "FreeMemory", Type: "gauge", Value: int(free)},
		)
	}

	cpuValue := ms.cpuUtilization
	ms.metrics.MetricList = append(ms.metrics.MetricList,
		models.Metric{Name: "CPUtilization", Type: "gauge", Value: int(cpuValue)},
	)
}

func (ms *MetricsService) addMetrics(m runtime.MemStats) {
	ms.metrics.MetricList = []models.Metric{
		{Type: models.Gauge, Name: "Alloc", Value: int(m.Alloc)},
		{Type: models.Gauge, Name: "BuckHashSys", Value: int(m.BuckHashSys)},
		{Type: models.Gauge, Name: "Frees", Value: int(m.Frees)},
		{Type: models.Gauge, Name: "GCCPUFraction", Value: int(m.GCCPUFraction)},
		{Type: models.Gauge, Name: "GCSys", Value: int(m.GCSys)},
		{Type: models.Gauge, Name: "HeapAlloc", Value: int(m.HeapAlloc)},
		{Type: models.Gauge, Name: "HeapIdle", Value: int(m.HeapIdle)},
		{Type: models.Gauge, Name: "HeapInuse", Value: int(m.HeapInuse)},
		{Type: models.Gauge, Name: "HeapObjects", Value: int(m.HeapObjects)},
		{Type: models.Gauge, Name: "HeapReleased", Value: int(m.HeapReleased)},
		{Type: models.Gauge, Name: "HeapSys", Value: int(m.HeapSys)},
		{Type: models.Gauge, Name: "LastGC", Value: int(m.LastGC)},
		{Type: models.Gauge, Name: "Lookups", Value: int(m.Lookups)},
		{Type: models.Gauge, Name: "MCacheInuse", Value: int(m.MCacheInuse)},
		{Type: models.Gauge, Name: "MCacheSys", Value: int(m.MCacheSys)},
		{Type: models.Gauge, Name: "MSpanInuse", Value: int(m.MSpanInuse)},
		{Type: models.Gauge, Name: "MSpanSys", Value: int(m.MSpanSys)},
		{Type: models.Gauge, Name: "Mallocs", Value: int(m.Mallocs)},
		{Type: models.Gauge, Name: "NextGC", Value: int(m.NextGC)},
		{Type: models.Gauge, Name: "NumForcedGC", Value: int(m.NumForcedGC)},
		{Type: models.Gauge, Name: "NumGC", Value: int(m.NumGC)},
		{Type: models.Gauge, Name: "OtherSys", Value: int(m.OtherSys)},
		{Type: models.Gauge, Name: "PauseTotalNs", Value: int(m.PauseTotalNs)},
		{Type: models.Gauge, Name: "StackInuse", Value: int(m.StackInuse)},
		{Type: models.Gauge, Name: "StackSys", Value: int(m.StackSys)},
		{Type: models.Gauge, Name: "Sys", Value: int(m.Sys)},
		{Type: models.Gauge, Name: "TotalAlloc", Value: int(m.TotalAlloc)},
		{Type: models.Gauge, Name: "RandomValue", Value: int(rand.Int63())},

		{Type: models.Counter, Name: "PollCount", Value: ms.metrics.PollCount},
	}
}

func (ms *MetricsService) GetMetrics() *models.Metrics {
	return ms.metrics
}
