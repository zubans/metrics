package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/zubans/metrics/internal/errdefs"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
)

// Example: обновление метрики через POST /update/{type}/{name}/{value}
func ExampleHandler_UpdateMetric() {
	h := handler.NewHandler(services.NewMetricService(storage.NewMemStorage()))
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.45", nil)
	w := httptest.NewRecorder()

	h.UpdateMetric(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

// Example: получение метрики через GET /value/{type}/{name}
func ExampleHandler_GetMetric() {
	h := handler.NewHandler(services.NewMetricService(storage.NewMemStorage()))
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	w := httptest.NewRecorder()

	h.GetMetric(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

// Example: обновление метрики через POST /update/ (JSON)
func ExampleHandler_UpdateMetricJSON() {
	h := handler.NewHandler(services.NewMetricService(storage.NewMemStorage()))
	metric := models.MetricsDTO{ID: "Alloc", MType: "gauge", Value: floatPtr(123.45)}
	body, _ := json.Marshal(metric)
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateMetricJSON(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

// Example: получение метрики через POST /value/ (JSON)
func ExampleHandler_GetMetricJSON() {
	h := handler.NewHandler(services.NewMetricService(storage.NewMemStorage()))
	metric := models.MetricsDTO{ID: "Alloc", MType: "gauge"}
	body, _ := json.Marshal(metric)
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GetMetricJSON(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

// Example: просмотр всех метрик через GET /
func ExampleHandler_ShowMetrics() {
	h := handler.NewHandler(services.NewMetricService(storage.NewMemStorage()))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ShowMetrics(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

// mockServicePing реализует ServerMetricService только для Ping.
type mockServicePing struct{}

func (m *mockServicePing) UpdateMetric(ctx context.Context, mData *services.MetricData) (*models.MetricsDTO, *errdefs.CustomError, error) {
	return nil, nil, nil
}
func (m *mockServicePing) UpdateMetrics(ctx context.Context, mData []models.MetricsDTO) (bool, *errdefs.CustomError, error) {
	return true, nil, nil
}
func (m *mockServicePing) GetMetric(ctx context.Context, mData *services.MetricData) (string, *errdefs.CustomError) {
	return "", nil
}
func (m *mockServicePing) GetJSONMetric(ctx context.Context, jsonData *models.MetricsDTO) ([]byte, *errdefs.CustomError) {
	return nil, nil
}
func (m *mockServicePing) ShowMetrics(ctx context.Context) (string, error) { return "", nil }
func (m *mockServicePing) Ping(ctx context.Context) error                  { return nil }

// Example: проверка доступности сервера через GET /ping
func ExampleHandler_PingServer() {
	h := handler.NewHandler(&mockServicePing{})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.PingServer(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	// Output:
}

func floatPtr(f float64) *float64 { return &f }
