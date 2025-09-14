package controllers

import (
	"log"

	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/transport"
)

type MetricControllerer interface {
	UpdateMetrics()
	SendMetrics()
}

type MetricsController struct {
	metricsService *services.MetricsService
	transport      transport.TransportInterface
}

func NewMetricsController(metricsService *services.MetricsService) *MetricsController {
	return &MetricsController{
		metricsService: metricsService,
		transport:      transport.NewTransport(metricsService),
	}
}

func (mc *MetricsController) UpdateMetrics() {
	mc.metricsService.CollectMetrics()
}

func (mc *MetricsController) SendMetrics() {
	metrics := mc.metricsService.GetMetrics()
	if err := mc.transport.SendMetrics(metrics.MetricList); err != nil {
		log.Printf("Error sending metrics: %v", err)
	}
}

func (mc *MetricsController) Close() {
	if err := mc.transport.Close(); err != nil {
		log.Printf("Error closing transport: %v", err)
	}
}

// GetMetricsService returns the metrics service for testing purposes
func (mc *MetricsController) GetMetricsService() *services.MetricsService {
	return mc.metricsService
}
