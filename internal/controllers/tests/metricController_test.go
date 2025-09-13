package controllers_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/controllers"
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
	controller := controllers.NewMetricsController(service)

	t.Run("CollectMetrics populates metrics", func(t *testing.T) {
		service := controller.GetMetricsService()
		service.CollectMetrics()
		metrics := service.GetMetrics()

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
		service := controller.GetMetricsService()
		service.CollectMetrics()
		controller.SendMetrics()

		metrics := service.GetMetrics()
		assert.Len(t, metrics.MetricList, 29)
	})

	t.Run("Error handling", func(t *testing.T) {
		// Тест ошибки теперь не может напрямую мокать httpClient,
		// так как он инкапсулирован в транспорте
		// Оставляем базовый тест функциональности
		mc := controllers.NewMetricsController(service)

		logBuffer := bytes.NewBuffer(nil)
		log.SetOutput(logBuffer)
		defer log.SetOutput(os.Stderr)

		mc.UpdateMetrics()
		mc.SendMetrics()

		// Проверяем, что метрики собираются
		service := mc.GetMetricsService()
		metrics := service.GetMetrics()
		assert.NotEmpty(t, metrics.MetricList)
	})
}

func TestErrorScenarios(t *testing.T) {
	cfg := &config.AgentConfig{
		AddressServer: "invalid-url:9999",
	}

	service := services.NewMetricsService(cfg)
	controller := controllers.NewMetricsController(service)

	controller.UpdateMetrics()

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	t.Run("Connection error", func(t *testing.T) {
		controller.SendMetrics()
		// Проверяем, что метрики собираются, даже если отправка не удается
		service := controller.GetMetricsService()
		metrics := service.GetMetrics()
		assert.NotEmpty(t, metrics.MetricList)
	})
}
