package controllers

import (
	"fmt"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"net/http"
)

type MetricControllerer interface {
	NewMetricsController(metricsService *services.MetricsService) *MetricsController
	UpdateMetrics()
	SendMetrics()
}

type MetricsController struct {
	metricsService *services.MetricsService
}

func NewMetricsController(metricsService *services.MetricsService) *MetricsController {
	return &MetricsController{
		metricsService: metricsService,
	}
}

func (mc *MetricsController) UpdateMetrics() {
	mc.metricsService.CollectMetrics()
}

func (mc *MetricsController) SendMetrics() {
	metrics := mc.metricsService.GetMetrics()

	for _, metric := range metrics.MetricList {
		url := ToURL(metric, mc.metricsService.Cfg)

		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Printf("Error sending metric %s: %v\n", metric.Name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Successfully sent metric: %s\n", metric.Name)
		} else {
			fmt.Printf("Failed to send metric: %s, status code: %d\n", metric.Name, resp.StatusCode)
		}
	}
}

func ToURL(m models.Metric, cfg *config.AgentConfig) string {

	return fmt.Sprintf("http://%s/update/%s/%s/%d", cfg.AddressServer, m.Type, m.Name, m.Value)
}
