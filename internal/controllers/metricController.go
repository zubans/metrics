package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"log"
	"net/http"
)

type MetricControllerer interface {
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
		url := fmt.Sprintf("http://%s/update/%s/%s/%d", mc.metricsService.Cfg.AddressServer, metric.Type, metric.Name, metric.Value)

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

func (mc *MetricsController) JSONSendMetrics() {
	metrics := mc.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/update/", mc.metricsService.Cfg.AddressServer)

	for _, metric := range dtoMetrics {
		b, _ := json.Marshal(metric)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
		if err != nil {
			log.Printf("Error sending metric %s: %v. BODY: %v\n", metric.ID, err, metric)
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			fmt.Printf("Failed to close Body: %s\n", err)
			return
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Successfully sent metric: %s\n", metric.ID)
		} else {
			fmt.Printf("Failed to send metric: %s, status code: %d\n", metric.ID, resp.StatusCode)
		}
	}
}
