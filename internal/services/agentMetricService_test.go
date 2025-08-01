package services

import (
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"testing"
)

func TestNewMetricsService(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)

	if service == nil {
		t.Fatal("NewMetricsService returned nil")
	}

	if service.Cfg != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, service.Cfg)
	}

	if service.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}

	if service.metrics.PollCount != 0 {
		t.Errorf("Expected initial PollCount to be 0, got %d", service.metrics.PollCount)
	}
}

func TestMetricsService_CollectMetrics(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)

	initialPollCount := service.metrics.PollCount
	if initialPollCount != 0 {
		t.Errorf("Expected initial PollCount to be 0, got %d", initialPollCount)
	}

	service.CollectMetrics()

	if service.metrics.PollCount != initialPollCount+1 {
		t.Errorf("Expected PollCount to be %d, got %d", initialPollCount+1, service.metrics.PollCount)
	}

	if len(service.metrics.MetricList) == 0 {
		t.Error("Expected metrics to be collected")
	}

	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue", "PollCount",
	}

	metricNames := make(map[string]bool)
	for _, metric := range service.metrics.MetricList {
		metricNames[metric.Name] = true
	}

	for _, expectedMetric := range expectedMetrics {
		if !metricNames[expectedMetric] {
			t.Errorf("Expected metric %s to be present", expectedMetric)
		}
	}
}

func TestMetricsService_GetMetrics(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)

	service.CollectMetrics()

	metrics := service.GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics returned nil")
	}

	if metrics.PollCount != 1 {
		t.Errorf("Expected PollCount to be 1, got %d", metrics.PollCount)
	}

	if len(metrics.MetricList) == 0 {
		t.Error("Expected metrics to be present")
	}

	if metrics != service.metrics {
		t.Error("GetMetrics should return the same metrics object")
	}
}

func TestMetricsService_CollectMetrics_MultipleCalls(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)

	for i := 0; i < 5; i++ {
		service.CollectMetrics()
	}

	expectedPollCount := 5
	if service.metrics.PollCount != expectedPollCount {
		t.Errorf("Expected PollCount to be %d, got %d", expectedPollCount, service.metrics.PollCount)
	}

	expectedMetricCount := 29 // 28 runtime метрик + PollCount
	if len(service.metrics.MetricList) != expectedMetricCount {
		t.Errorf("Expected %d metrics, got %d", expectedMetricCount, len(service.metrics.MetricList))
	}
}

func TestMetricsService_MetricsTypes(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)
	service.CollectMetrics()

	gaugeCount := 0
	counterCount := 0

	for _, metric := range service.metrics.MetricList {
		switch metric.Type {
		case models.Gauge:
			gaugeCount++
		case models.Counter:
			counterCount++
		default:
			t.Errorf("Unexpected metric type: %s", metric.Type)
		}
	}

	if gaugeCount == 0 {
		t.Error("Expected gauge metrics to be present")
	}

	if counterCount == 0 {
		t.Error("Expected counter metrics to be present")
	}

	pollCountFound := false
	for _, metric := range service.metrics.MetricList {
		if metric.Name == "PollCount" && metric.Type == models.Counter {
			pollCountFound = true
			break
		}
	}

	if !pollCountFound {
		t.Error("Expected PollCount to be a counter metric")
	}
}

func TestMetricsService_RandomValue(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)
	service.CollectMetrics()

	randomValueFound := false
	for _, metric := range service.metrics.MetricList {
		if metric.Name == "RandomValue" && metric.Type == models.Gauge {
			randomValueFound = true
			break
		}
	}

	if !randomValueFound {
		t.Error("Expected RandomValue metric to be present")
	}
}

func TestMetricsService_InterfaceCompliance(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	var collector MetricsCollector = NewMetricsService(cfg)

	if collector == nil {
		t.Fatal("MetricsService should implement MetricsCollector interface")
	}

	collector.CollectMetrics()
	metrics := collector.GetMetrics()

	if metrics == nil {
		t.Error("GetMetrics should not return nil")
	}
}

func TestMetricsService_ConcurrentAccess(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := NewMetricsService(cfg)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			service.CollectMetrics()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	expectedPollCount := 10
	if service.metrics.PollCount != expectedPollCount {
		t.Errorf("Expected PollCount to be %d, got %d", expectedPollCount, service.metrics.PollCount)
	}
}
