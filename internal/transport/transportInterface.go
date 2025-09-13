package transport

import "github.com/zubans/metrics/internal/models"

type TransportInterface interface {
	SendMetrics(metrics []models.Metric) error
	SendSingleMetric(metric models.Metric) error
	Close() error
}
