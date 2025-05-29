package controllers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMetricsController_JSONSendMetrics(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "localhost:8080",
		PollInterval:  2,
		SendInterval:  10,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		var metric []models.MetricsDTO
		err = json.Unmarshal(body, &metric)
		require.NoError(t, err)

		assert.Contains(t, []string{"gauge", "counter"}, metric[0].MType)
		assert.NotEmpty(t, metric[0].ID)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg.AddressServer = server.URL[7:] //убираем "http://"

	service := services.NewMetricsService(cfg)
	controller := &MetricsController{
		metricsService: service,
		httpClient:     resty.New(),
	}

	t.Run("CollectMetrics populates metrics", func(t *testing.T) {
		controller.metricsService.CollectMetrics()
		metrics := controller.metricsService.GetMetrics()

		assert.NotEmpty(t, metrics.MetricList)
		assert.Greater(t, metrics.PollCount, 0)
		expectedType := models.Gauge
		expectedName := "Alloc"

		found := false
		for _, m := range metrics.MetricList {
			if m.Type == expectedType && m.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Metric %s/%s not found", expectedType, expectedName)
	})

	t.Run("Successful metrics sending", func(t *testing.T) {
		controller.metricsService.CollectMetrics()
		controller.JSONSendMetrics()

		metrics := controller.metricsService.GetMetrics()
		assert.Len(t, metrics.MetricList, 32)
	})

	t.Run("Error handling", func(t *testing.T) {
		mc := NewMetricsController(service)
		mc.httpClient.
			SetTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("connection refused")
			}))

		logBuffer := bytes.NewBuffer(nil)
		log.SetOutput(logBuffer)
		defer log.SetOutput(os.Stderr)

		mc.UpdateMetrics()
		mc.JSONSendMetrics()

		require.Contains(t, logBuffer.String(), "Error sending metric",
			"Должна быть ошибка отправки метрики")
		require.Contains(t, logBuffer.String(), "connection refused",
			"В логе должна быть причина ошибки")
	})
}

func TestErrorScenarios(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "invalid-url:9999",
	}

	service := services.NewMetricsService(cfg)
	controller := &MetricsController{
		metricsService: service,
		httpClient:     resty.New(),
	}

	controller.UpdateMetrics()

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	t.Run("Connection error", func(t *testing.T) {
		controller.JSONSendMetrics()
		assert.Contains(t, logBuffer.String(), "Error sending metric")
	})
}
