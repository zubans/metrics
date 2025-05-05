package controllers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"io"
	"log"
	"net/http"
	"time"
)

var gzipNewWriter = func(w io.Writer) *gzip.Writer {
	return gzip.NewWriter(w)
}

type MetricControllerer interface {
	UpdateMetrics()
	SendMetrics()
}

type MetricsController struct {
	metricsService *services.MetricsService
	httpClient     *http.Client
}

func NewMetricsController(metricsService *services.MetricsService) *MetricsController {
	return &MetricsController{
		metricsService: metricsService,
		httpClient: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
	}
}

func (mc *MetricsController) UpdateMetrics() {
	mc.metricsService.CollectMetrics()
}

func (mc *MetricsController) JSONSendMetrics() {
	metrics := mc.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/updates/", mc.metricsService.Cfg.AddressServer)

	body, err := json.Marshal(dtoMetrics)
	var buf bytes.Buffer

	gz := gzipNewWriter(&buf)

	_, err = gz.Write(body)
	if err != nil {
		log.Println("Error compressing metric data")
		return
	}

	err = gz.Close()
	if err != nil {
		log.Println("Error close gzip compressor")
		return
	}

	req, _ := http.NewRequest("POST", url, &buf)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	response, err := mc.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending metric: %v. BODY: %v\n", err, metrics)
		return
	}

	err = response.Body.Close()
	if err != nil {
		log.Printf("Failed to close Body: %s\n", err)
		return
	}

	if response.StatusCode == http.StatusOK {
		log.Printf("Successfully sent metric: %v\n", metrics)
	} else {
		log.Printf("Failed to send metric: %v, status code: %d\n", metrics, response.StatusCode)
	}
}
