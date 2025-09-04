package controllers

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/zubans/metrics/internal/cryptoutil"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
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
	publicKey      *rsa.PublicKey
}

func NewMetricsController(metricsService *services.MetricsService) *MetricsController {
	mc := &MetricsController{
		metricsService: metricsService,
		httpClient: resty.New().
			SetTimeout(3 * time.Second),
	}
	if metricsService.Cfg != nil && metricsService.Cfg.CryptoKey != "" {
		if pub, err := cryptoutil.LoadPublicKey(metricsService.Cfg.CryptoKey); err == nil {
			mc.publicKey = pub
		} else {
			log.Printf("failed to load public key: %v", err)
		}
	}
	return mc
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

	reqBody, extraHeaders, err := mc.prepareRequestBody(buf.Bytes())
	if err != nil {
		log.Printf("encryption failed: %v", err)
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
		SetHeader("X-Real-IP", detectHostIP())

	if mc.publicKey == nil {
		request = request.SetHeader("Content-Encoding", "gzip")
	}
	for k, v := range extraHeaders {
		if v == "" {
			continue
		}
		request = request.SetHeader(k, v)
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

func (mc *MetricsController) prepareRequestBody(data []byte) (interface{}, map[string]string, error) {
	extraHeaders := map[string]string{}
	if mc.publicKey != nil {
		env, encErr := cryptoutil.EncryptHybrid(mc.publicKey, data)
		if encErr != nil {
			return nil, nil, encErr
		}
		extraHeaders["X-Encrypted"] = "1"
		extraHeaders["Content-Encoding"] = ""
		return env, extraHeaders, nil
	}
	return data, extraHeaders, nil
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

		var reqBody interface{}
		var extraHeaders = map[string]string{}
		if mc.publicKey != nil {
			env, encErr := cryptoutil.EncryptHybrid(mc.publicKey, buf.Bytes())
			if encErr != nil {
				log.Printf("encryption failed: %v", encErr)
				continue
			}
			reqBody = env
			extraHeaders["X-Encrypted"] = "1"
		} else {
			reqBody = buf.Bytes()
		}

		restyReq := mc.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Real-IP", detectHostIP())
		if mc.publicKey == nil {
			restyReq = restyReq.SetHeader("Content-Encoding", "gzip")
		}
		for k, v := range extraHeaders {
			if v == "" {
				continue
			}
			restyReq = restyReq.SetHeader(k, v)
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
