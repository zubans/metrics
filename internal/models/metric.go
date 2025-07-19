package models

// MetricType определяет тип метрики: gauge или counter.
type MetricType string

const (
	// Gauge — метрика типа gauge.
	Gauge MetricType = "gauge"
	// Counter — метрика типа counter.
	Counter MetricType = "counter"
)

// Metric описывает одну метрику.
type Metric struct {
	Type  MetricType
	Name  string
	Value int
}

// Metrics содержит список метрик и счётчик PollCount.
type Metrics struct {
	MetricList []Metric
	PollCount  int
}

// MetricsDTO — структура для передачи метрик в формате JSON.
type MetricsDTO struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// ConvertToDTO преобразует Metric в MetricsDTO.
func ConvertToDTO(m Metric) MetricsDTO {
	dto := MetricsDTO{
		ID:    m.Name,
		MType: string(m.Type),
	}

	switch m.Type {
	case Gauge:
		val := float64(m.Value)
		dto.Value = &val
	case Counter:
		delta := int64(m.Value)
		dto.Delta = &delta
	}

	return dto
}

// ConvertMetricsListToDTO преобразует список Metric в список MetricsDTO.
func ConvertMetricsListToDTO(metrics []Metric) []MetricsDTO {
	result := make([]MetricsDTO, 0, len(metrics))
	for _, m := range metrics {
		result = append(result, ConvertToDTO(m))
	}
	return result
}
