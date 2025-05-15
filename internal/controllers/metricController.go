package controllers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
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
	httpClient     *resty.Client
}

func NewMetricsController(metricsService *services.MetricsService) *MetricsController {
	return &MetricsController{
		metricsService: metricsService,
		httpClient: resty.New().
			SetTimeout(3 * time.Second),
	}
}

func (mc *MetricsController) UpdateMetrics() {
	mc.metricsService.CollectMetrics()
}

func (mc *MetricsController) JSONSendMetrics() {
	retryDelays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	metrics := mc.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/updates/", mc.metricsService.Cfg.AddressServer)

	body, err := json.Marshal(dtoMetrics)
	if err != nil {
		log.Println("Error json Encode metric data")
		return
	}

	var buf bytes.Buffer
	var hash []byte

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

	request := mc.httpClient.
		SetRetryCount(3).
		SetRetryWaitTime(1*time.Second).
		SetRetryMaxWaitTime(5*time.Second).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			attempt := r.Request.Attempt - 1
			if attempt >= len(retryDelays) {
				attempt = len(retryDelays) - 1
			}

			return retryDelays[attempt], nil
		}).
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes())

	if mc.metricsService.Cfg.Key != "" {
		h := hmac.New(sha256.New, []byte(mc.metricsService.Cfg.Key))
		h.Write(body)
		hash = h.Sum(nil)
		request.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}

	response, err := request.Post(url)

	if err != nil {
		log.Printf("Error sending metric: %v. BODY: %v. HASH: %v\n", err, metrics, hex.EncodeToString(hash))
		return
	}

	if response.IsSuccess() {
		log.Printf("Successfully sent metric: %v\n", metrics)
	} else {
		log.Printf("Failed to send metric: %v, status code: %d\n", metrics, response.StatusCode())
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

		response, err := mc.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(buf.Bytes()).
			Post(url)
		if err != nil {
			log.Printf("Error sending metric %s: %v. BODY: %v\n", metric.ID, err, metric)
			continue
		}

		if response.IsSuccess() {
			log.Printf("Successfully sent metric: %s\n", metric.ID)
		} else {
			log.Printf("Failed to send metric: %s, status code: %d\n", metric.ID, response.StatusCode())
		}
	}
}
