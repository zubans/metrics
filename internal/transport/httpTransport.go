package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
)

type HTTPTransport struct {
	metricsService *services.MetricsService
	httpClient     *resty.Client
	dataProcessor  *DataProcessor
}

func NewHTTPTransport(metricsService *services.MetricsService) *HTTPTransport {
	httpClient := resty.New().SetTimeout(metricsService.Cfg.HTTPTimeout)

	return &HTTPTransport{
		metricsService: metricsService,
		httpClient:     httpClient,
		dataProcessor:  NewDataProcessor(metricsService),
	}
}

func (ht *HTTPTransport) SendMetrics(metrics []models.Metric) error {
	ht.sendMetricsHTTP()
	return nil
}

func (ht *HTTPTransport) SendSingleMetric(metric models.Metric) error {
	ht.sendSingleMetricsHTTP()
	return nil
}

func (ht *HTTPTransport) Close() error {
	return nil
}

func (ht *HTTPTransport) sendMetricsHTTP() {
	retryDelays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	metrics := ht.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/updates/", ht.metricsService.Cfg.AddressServer)

	body, err := json.Marshal(dtoMetrics)
	if err != nil {
		log.Println("Error json Encode metric data")
		return
	}

	reqBody, extraHeaders, err := ht.dataProcessor.ProcessData(body)
	if err != nil {
		log.Printf("Error processing data: %v", err)
		return
	}

	request := ht.httpClient.
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
		SetHeader("X-Real-IP", detectHostIP())

	for k, v := range extraHeaders {
		if v != "" {
			request = request.SetHeader(k, v)
		}
	}
	response, err := request.SetBody(reqBody).Post(url)

	if err != nil {
		log.Printf("Error sending metric: %v. BODY: %v\n", err, metrics)
		return
	}

	if response.IsSuccess() {
		log.Printf("Successfully sent metric: %v\n", metrics)
	} else {
		log.Printf("Failed to send metric: %v, status code: %d\n", metrics, response.StatusCode())
	}
}

func (ht *HTTPTransport) sendSingleMetricsHTTP() {
	metrics := ht.metricsService.GetMetrics()
	dtoMetrics := models.ConvertMetricsListToDTO(metrics.MetricList)

	url := fmt.Sprintf("http://%s/update/", ht.metricsService.Cfg.AddressServer)

	for _, metric := range dtoMetrics {
		b, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Error json encode metric data")
			continue
		}

		reqBody, extraHeaders, err := ht.dataProcessor.ProcessData(b)
		if err != nil {
			log.Printf("Error processing data: %v", err)
			continue
		}

		restyReq := ht.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Real-IP", detectHostIP())

		for k, v := range extraHeaders {
			if v != "" {
				restyReq = restyReq.SetHeader(k, v)
			}
		}

		response, err := restyReq.
			SetBody(reqBody).
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

func detectHostIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue
		}
		return ip.String()
	}
	return ""
}
