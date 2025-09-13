package services_test

import (
	"testing"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
)

func TestNewMetricsService(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := services.NewMetricsService(cfg)

	if service == nil {
		t.Fatal("NewMetricsService returned nil")
	}

	if service.Cfg != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, service.Cfg)
	}

	metrics := service.GetMetricsInternal()
	if metrics == nil {
		t.Error("Expected metrics to be initialized")
	}

	if metrics.PollCount != 0 {
		t.Errorf("Expected initial PollCount to be 0, got %d", metrics.PollCount)
	}
}

func TestMetricsService_CollectMetrics(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := services.NewMetricsService(cfg)

	metrics := service.GetMetricsInternal()
	initialPollCount := metrics.PollCount
	if initialPollCount != 0 {
		t.Errorf("Expected initial PollCount to be 0, got %d", initialPollCount)
	}

	service.CollectMetrics()

	if metrics.PollCount != initialPollCount+1 {
		t.Errorf("Expected PollCount to be %d, got %d", initialPollCount+1, metrics.PollCount)
	}

	if len(metrics.MetricList) == 0 {
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
	for _, metric := range metrics.MetricList {
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

	service := services.NewMetricsService(cfg)

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

	internalMetrics := service.GetMetricsInternal()
	if metrics != internalMetrics {
		t.Error("GetMetrics should return the same metrics object")
	}
}

func TestMetricsService_CollectMetrics_MultipleCalls(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := services.NewMetricsService(cfg)

	for i := 0; i < 5; i++ {
		service.CollectMetrics()
	}

	metrics := service.GetMetricsInternal()
	expectedPollCount := 5
	if metrics.PollCount != expectedPollCount {
		t.Errorf("Expected PollCount to be %d, got %d", expectedPollCount, metrics.PollCount)
	}

	expectedMetricCount := 29 // 28 runtime метрик + PollCount
	if len(metrics.MetricList) != expectedMetricCount {
		t.Errorf("Expected %d metrics, got %d", expectedMetricCount, len(metrics.MetricList))
	}
}

func TestMetricsService_MetricsTypes(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10,
		PollInterval:  2,
	}

	service := services.NewMetricsService(cfg)
	service.CollectMetrics()

	metrics := service.GetMetricsInternal()
	gaugeCount := 0
	counterCount := 0

	for _, metric := range metrics.MetricList {
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
	for _, metric := range metrics.MetricList {
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

	service := services.NewMetricsService(cfg)
	service.CollectMetrics()

	metrics := service.GetMetricsInternal()
	randomValueFound := false
	for _, metric := range metrics.MetricList {
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

	var collector services.MetricsCollector = services.NewMetricsService(cfg)

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

	service := services.NewMetricsService(cfg)

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

	metrics := service.GetMetricsInternal()
	expectedPollCount := 10
	if metrics.PollCount != expectedPollCount {
		t.Errorf("Expected PollCount to be %d, got %d", expectedPollCount, metrics.PollCount)
	}
}
