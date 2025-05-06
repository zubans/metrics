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
	"strings"
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
	maxRetries := 3
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	metrics := mc.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/updates/", mc.metricsService.Cfg.AddressServer)

	body, err := json.Marshal(dtoMetrics)
	if err != nil {
		log.Println("Error json Encode metric data")
		return
	}

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

	var response *http.Response
	for trying := 0; trying <= maxRetries; trying++ {
		response, err = mc.httpClient.Do(req)
		if err != nil {
			if trying < maxRetries && strings.Contains(err.Error(), "connection refused") {
				log.Printf("Bad %v trying sending metric: %v. BODY: %v\n", trying+1, err, metrics)
				time.Sleep(retryDelays[trying])
				continue
			}
			log.Printf("Error sending metric: %v. BODY: %v\n", err, metrics)
			return
		}
		response.Body.Close()
	}

	if response != nil {
		if response.StatusCode == http.StatusOK {
			log.Printf("Successfully sent metric: %v\n", metrics)
		} else {
			log.Printf("Failed to send metric: %v, status code: %d\n", metrics, response.StatusCode)
		}
	}
}

func (mc *MetricsController) OldJSONSendMetrics() {
	metrics := mc.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/update/", mc.metricsService.Cfg.AddressServer)

	for _, metric := range dtoMetrics {
		b, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Error json encode metric data")
		}

		var buf bytes.Buffer

		gz := gzipNewWriter(&buf)

		_, err = gz.Write(b)
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
			log.Printf("Error sending metric %s: %v. BODY: %v\n", metric.ID, err, metric)
			continue
		}

		err = response.Body.Close()
		if err != nil {
			log.Printf("Failed to close Body: %s\n", err)
			return
		}

		if response.StatusCode == http.StatusOK {
			log.Printf("Successfully sent metric: %s\n", metric.ID)
		} else {
			log.Printf("Failed to send metric: %s, status code: %d\n", metric.ID, response.StatusCode)
		}
	}
}
