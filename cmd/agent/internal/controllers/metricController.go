package controllers

import (
	"fmt"
	"github.com/zubans/metrics/cmd/agent/internal/helpers"
	"github.com/zubans/metrics/cmd/agent/internal/services"
	"net/http"
)

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
		url := helpers.ToURL(metric)

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
