package handler

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_UpdateMetricJSON(t *testing.T) {
	newMemStorage := storage.NewMemStorage()
	newService := services.NewMetricService(newMemStorage)
	handler := NewHandler(newService)
	tests := []struct {
		name                string
		requestData         string
		expectedStatusCode  int
		expectedContentType string
	}{
		{
			name:                "Valid Counter Metric",
			requestData:         `{  "id": "PollCount",  "type": "counter",  "delta": 1}`,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json",
		},
		{
			name:                "Valid Gauge Metric",
			requestData:         `{  "id": "Alloc",  "type": "gauge",  "value": 1}`,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json",
		},
		{
			name:                "Invalid Gauge Metric - bad value type",
			requestData:         `{  "id": "Alloc",  "type": "gauge",  "value": "1""}`,
			expectedStatusCode:  http.StatusBadRequest,
			expectedContentType: "application/json",
		},
		{
			name:                "Invalid Gauge Metric - unsupported type",
			requestData:         `{  "id": "Alloc",  "type": "unsupported",  "value": 1"}`,
			expectedStatusCode:  http.StatusBadRequest,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/update/", bytes.NewBufferString(tt.requestData))
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h := http.HandlerFunc(handler.UpdateMetricJSON)
			h.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedContentType, rr.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_UpdateMetric(t *testing.T) {
	newMemStorage := storage.NewMemStorage()
	newService := services.NewMetricService(newMemStorage)
	handler := NewHandler(newService)

	tests := []struct {
		name               string
		metricType         string
		metricName         string
		metricValue        string
		expectedStatusCode int
		expectedGauge      float64
		expectedCounter    int64
	}{
		{
			name:               "Valid Gauge Metric",
			metricType:         "gauge",
			metricName:         "cpu_usage",
			metricValue:        "50.567",
			expectedStatusCode: http.StatusOK,
			expectedGauge:      50.567,
			expectedCounter:    0,
		},
		{
			name:               "Valid Counter Metric",
			metricType:         "counter",
			metricName:         "requests_count",
			metricValue:        "100",
			expectedStatusCode: http.StatusOK,
			expectedGauge:      0,
			expectedCounter:    100,
		},
		{
			name:               "Invalid Metric Type",
			metricType:         "invalid",
			metricName:         "invalid_metric",
			metricValue:        "100",
			expectedStatusCode: http.StatusBadRequest,
			expectedGauge:      0,
			expectedCounter:    0,
		},
		{
			name:               "Invalid Gauge Value",
			metricType:         "gauge",
			metricName:         "cpu_usage",
			metricValue:        "invalid_value",
			expectedStatusCode: http.StatusBadRequest,
			expectedGauge:      0,
			expectedCounter:    0,
		},
		{
			name:               "Invalid Counter Value",
			metricType:         "counter",
			metricName:         "requests_count",
			metricValue:        "invalid_value",
			expectedStatusCode: http.StatusBadRequest,
			expectedGauge:      0,
			expectedCounter:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/update/"+test.metricType+"/"+test.metricName+"/"+test.metricValue, nil)
			if err != nil {
				t.Fatal(err)
			}

			newRecorder := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)

			r.ServeHTTP(newRecorder, req)

			assert.Equal(t, test.expectedStatusCode, newRecorder.Code)

			if test.expectedStatusCode == http.StatusOK {
				if test.metricType == "gauge" {
					gaugeValue, exists := newMemStorage.GetGauge(context.Background(), test.metricName)
					if assert.True(t, exists, "Expected  gauge value is exist") {
						assert.Equal(t, test.expectedGauge, gaugeValue, "Gauge value mismatch")
					}
				} else if test.metricType == "counter" {
					counterValue, exists := newMemStorage.GetCounter(context.Background(), test.metricName)
					if assert.True(t, exists, "Expected counter value is exist") {
						assert.Equal(t, test.expectedCounter, counterValue, "Counter value mismatch")
					}
				}
			}
		})
	}
}
