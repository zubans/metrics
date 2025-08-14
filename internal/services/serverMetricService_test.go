package services

import (
	"context"
	"github.com/zubans/metrics/internal/models"
	"testing"
)

type MockMetricStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMockMetricStorage() *MockMetricStorage {
	return &MockMetricStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MockMetricStorage) UpdateGauge(_ context.Context, name string, value float64) float64 {
	m.gauges[name] = value
	return value
}

func (m *MockMetricStorage) UpdateCounter(_ context.Context, name string, value int64) int64 {
	if existing, exists := m.counters[name]; exists {
		m.counters[name] = existing + value
		return m.counters[name]
	}
	m.counters[name] = value
	return value
}

func (m *MockMetricStorage) GetGauge(_ context.Context, name string) (float64, bool) {
	value, exists := m.gauges[name]
	return value, exists
}

func (m *MockMetricStorage) GetCounter(_ context.Context, name string) (int64, bool) {
	value, exists := m.counters[name]
	return value, exists
}

func (m *MockMetricStorage) ShowMetrics(_ context.Context) (map[string]float64, map[string]int64, error) {
	return m.gauges, m.counters, nil
}

func (m *MockMetricStorage) UpdateMetrics(ctx context.Context, metrics []models.MetricsDTO) error {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				m.gauges[metric.ID] = *metric.Value
			}
		case "counter":
			if metric.Delta != nil {
				m.UpdateCounter(ctx, metric.ID, *metric.Delta)
			}
		}
	}
	return nil
}

func TestNewMetricService(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	if service == nil {
		t.Fatal("NewMetricService returned nil")
	}

	if service.storage != mockStorage {
		t.Error("Expected storage to be set correctly")
	}
}

func TestNewMetricData(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		value       string
		expectError bool
	}{
		{
			name:        "valid gauge metric",
			metricType:  "gauge",
			metricName:  "test_metric",
			value:       "123.45",
			expectError: false,
		},
		{
			name:        "valid counter metric",
			metricType:  "counter",
			metricName:  "test_counter",
			value:       "100",
			expectError: false,
		},
		{
			name:        "invalid metric type",
			metricType:  "invalid",
			metricName:  "test_metric",
			value:       "123.45",
			expectError: true,
		},
		{
			name:        "empty metric name",
			metricType:  "gauge",
			metricName:  "",
			value:       "123.45",
			expectError: true,
		},
		{
			name:        "valid metric without value",
			metricType:  "gauge",
			metricName:  "test_metric",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metricData *MetricData
			var err error

			if tt.value != "" {
				metricData, err = NewMetricData(tt.metricType, tt.metricName, tt.value)
			} else {
				metricData, err = NewMetricData(tt.metricType, tt.metricName)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if metricData == nil {
				t.Error("Expected MetricData but got nil")
				return
			}

			if metricData.Type != tt.metricType {
				t.Errorf("Expected type %s, got %s", tt.metricType, metricData.Type)
			}

			if metricData.Name != tt.metricName {
				t.Errorf("Expected name %s, got %s", tt.metricName, metricData.Name)
			}

			if tt.value != "" {
				if metricData.Value == nil {
					t.Error("Expected value to be set")
				} else if *metricData.Value != tt.value {
					t.Errorf("Expected value %s, got %s", tt.value, *metricData.Value)
				}
			}
		})
	}
}

func TestParseMetricValue(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectValue float64
		expectError bool
	}{
		{
			name:        "valid integer",
			value:       "123",
			expectValue: 123.0,
			expectError: false,
		},
		{
			name:        "valid float",
			value:       "123.45",
			expectValue: 123.45,
			expectError: false,
		},
		{
			name:        "invalid value",
			value:       "invalid",
			expectValue: 0,
			expectError: true,
		},
		{
			name:        "empty value",
			value:       "",
			expectValue: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricData := &MetricData{
				Type:  "gauge",
				Name:  "test",
				Value: &tt.value,
			}

			value, err := ParseMetricValue(metricData)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value != tt.expectValue {
				t.Errorf("Expected value %f, got %f", tt.expectValue, value)
			}
		})
	}
}

func TestParseMetricValue_NilValue(t *testing.T) {
	metricData := &MetricData{
		Type: "gauge",
		Name: "test",
	}

	_, err := ParseMetricValue(metricData)
	if err == nil {
		t.Error("Expected error for nil value")
	}
}

func TestStorage_ShowMetrics(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	mockStorage.UpdateGauge(context.Background(), "test_gauge", 123.45)
	mockStorage.UpdateCounter(context.Background(), "test_counter", 100)

	result, err := service.ShowMetrics(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result, "<html>") {
		t.Error("Expected HTML in result")
	}

	if !contains(result, "Guages") {
		t.Error("Expected 'Guages' section in result")
	}

	if !contains(result, "Counters") {
		t.Error("Expected 'Counters' section in result")
	}
}

func TestStorage_GetMetric(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	mockStorage.UpdateGauge(context.Background(), "test_gauge", 123.45)
	mockStorage.UpdateCounter(context.Background(), "test_counter", 100)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		expectValue string
		expectError bool
	}{
		{
			name:        "existing gauge",
			metricType:  "gauge",
			metricName:  "test_gauge",
			expectValue: "123.45",
			expectError: false,
		},
		{
			name:        "existing counter",
			metricType:  "counter",
			metricName:  "test_counter",
			expectValue: "100",
			expectError: false,
		},
		{
			name:        "non-existing gauge",
			metricType:  "gauge",
			metricName:  "non_existing",
			expectValue: "",
			expectError: true,
		},
		{
			name:        "invalid metric type",
			metricType:  "invalid",
			metricName:  "test_gauge",
			expectValue: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricData := &MetricData{
				Type: tt.metricType,
				Name: tt.metricName,
			}

			result, customErr := service.GetMetric(context.Background(), metricData)

			if tt.expectError {
				if customErr == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if customErr != nil {
				t.Errorf("Unexpected error: %v", customErr)
				return
			}

			if result != tt.expectValue {
				t.Errorf("Expected value %s, got %s", tt.expectValue, result)
			}
		})
	}
}

func TestStorage_GetJSONMetric(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	mockStorage.UpdateGauge(context.Background(), "test_gauge", 123.45)
	mockStorage.UpdateCounter(context.Background(), "test_counter", 100)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		expectError bool
	}{
		{
			name:        "existing gauge",
			metricType:  "gauge",
			metricName:  "test_gauge",
			expectError: false,
		},
		{
			name:        "existing counter",
			metricType:  "counter",
			metricName:  "test_counter",
			expectError: false,
		},
		{
			name:        "non-existing metric",
			metricType:  "gauge",
			metricName:  "non_existing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := &models.MetricsDTO{
				ID:    tt.metricName,
				MType: tt.metricType,
			}

			result, customErr := service.GetJSONMetric(context.Background(), jsonData)

			if tt.expectError {
				if customErr == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if customErr != nil {
				t.Errorf("Unexpected error: %v", customErr)
				return
			}

			if result == nil {
				t.Error("Expected JSON result but got nil")
			}

			if len(result) == 0 {
				t.Error("Expected non-empty JSON result")
			}
		})
	}
}

func TestStorage_UpdateMetric(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		value       string
		expectError bool
	}{
		{
			name:        "valid gauge update",
			metricType:  "gauge",
			metricName:  "test_gauge",
			value:       "123.45",
			expectError: false,
		},
		{
			name:        "valid counter update",
			metricType:  "counter",
			metricName:  "test_counter",
			value:       "100",
			expectError: false,
		},
		{
			name:        "empty metric name",
			metricType:  "gauge",
			metricName:  "",
			value:       "123.45",
			expectError: true,
		},
		{
			name:        "missing gauge value",
			metricType:  "gauge",
			metricName:  "test_gauge",
			value:       "",
			expectError: true,
		},
		{
			name:        "missing counter value",
			metricType:  "counter",
			metricName:  "test_counter",
			value:       "",
			expectError: true,
		},
		{
			name:        "invalid metric type",
			metricType:  "invalid",
			metricName:  "test_metric",
			value:       "123.45",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricData := &MetricData{
				Type: tt.metricType,
				Name: tt.metricName,
			}

			if tt.value != "" {
				metricData.Value = &tt.value
			}

			result, customErr, err := service.UpdateMetric(context.Background(), metricData)

			if tt.expectError {
				if customErr == nil && err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if customErr != nil {
				t.Errorf("Unexpected custom error: %v", customErr)
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			if result.ID != tt.metricName {
				t.Errorf("Expected ID %s, got %s", tt.metricName, result.ID)
			}

			if result.MType != tt.metricType {
				t.Errorf("Expected type %s, got %s", tt.metricType, result.MType)
			}
		})
	}
}

func TestStorage_UpdateMetrics(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	tests := []struct {
		name        string
		metrics     []models.MetricsDTO
		expectError bool
	}{
		{
			name: "valid metrics update",
			metrics: []models.MetricsDTO{
				{ID: "test_gauge", MType: "gauge", Value: float64Ptr(123.45)},
				{ID: "test_counter", MType: "counter", Delta: int64Ptr(100)},
			},
			expectError: false,
		},
		{
			name:        "nil metrics",
			metrics:     nil,
			expectError: true,
		},
		{
			name:        "empty metrics slice",
			metrics:     []models.MetricsDTO{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success, customErr, err := service.UpdateMetrics(context.Background(), tt.metrics)

			if tt.expectError {
				if customErr == nil && err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if customErr != nil {
				t.Errorf("Unexpected custom error: %v", customErr)
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !success {
				t.Error("Expected success but got false")
			}
		})
	}
}

func TestStorage_InterfaceCompliance(t *testing.T) {
	mockStorage := NewMockMetricStorage()
	service := NewMetricService(mockStorage)

	if service == nil {
		t.Fatal("NewMetricService should not return nil")
	}

	ctx := context.Background()

	result, err := service.ShowMetrics(ctx)
	if err != nil {
		t.Errorf("ShowMetrics failed: %v", err)
	}
	if result == "" {
		t.Error("ShowMetrics should return non-empty result")
	}

	metricData := &MetricData{
		Type:  "gauge",
		Name:  "test_metric",
		Value: stringPtr("123.45"),
	}

	resultDTO, customErr, err := service.UpdateMetric(ctx, metricData)
	if err != nil {
		t.Errorf("UpdateMetric failed: %v", err)
	}
	if customErr != nil {
		t.Errorf("UpdateMetric custom error: %v", customErr)
	}
	if resultDTO == nil {
		t.Error("UpdateMetric should return result")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:], substr)))
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
