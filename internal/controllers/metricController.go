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

	var hash []byte

	buf, err := mc.compressData(body)
	if err != nil {
		log.Println("Error compress metric data")
		return
	}

	req := mc.httpClient.
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
		h := mc.calculateHash(body, mc.metricsService.Cfg.Key)
		req.SetHeader("HashSHA256", h)
	}

	response, err := req.Post(url)

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

func (mc *MetricsController) compressData(body []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(body); err != nil {
		log.Println("Error compressing metric data:", err)
		return nil, err
	}

	if err := gz.Close(); err != nil {
		log.Println("Error closing gzip compressor:", err)
		return nil, err
	}

	return &buf, nil
}

func (mc *MetricsController) calculateHash(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
