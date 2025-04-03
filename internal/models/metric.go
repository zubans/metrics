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
