package models

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Name  string
	Value int
}

type Metrics struct {
	MetricList []Metric
	PollCount  int
}

type MetricsDTO struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

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

func ConvertMetricsListToDTO(metrics []Metric) []MetricsDTO {
	result := make([]MetricsDTO, 0, len(metrics))
	for _, m := range metrics {
		result = append(result, ConvertToDTO(m))
	}
	return result
}
