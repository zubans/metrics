package controllers

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"io"
	"net/http"
	"testing"
)

var httpPost = http.Post

func TestMetricsController_SendMetrics(t *testing.T) {
	cfg := config.NewAgentConfig()
	type fields struct {
		metricsService *services.MetricsService
	}
	tests := []struct {
		name      string
		fields    fields
		response  *http.Response
		postError error
		url       func(models.Metric) string
	}{
		{
			name: "success",
			fields: fields{
				metricsService: services.NewMetricsService(cfg),
			},
			response: &http.Response{
				StatusCode: http.StatusOK,
			},
			postError: nil,
			url: func(metric models.Metric) string {

				return fmt.Sprintf("http://%s/update/%s/%s/%d", cfg.AddressServer, metric.Type, metric.Name, metric.Value)
			},
		},
		{
			name: "error",
			fields: fields{
				metricsService: services.NewMetricsService(cfg),
			},
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			postError: fmt.Errorf("error sending metric"),
			url: func(metric models.Metric) string {
				cfg := &config.AgentConfig{
					AddressServer: "localhost:8080",
				}
				return fmt.Sprintf("http://%s/update/%s/%s/%d", cfg.AddressServer, metric.Type, metric.Name, metric.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpPost = func(url string, contentType string, body io.Reader) (*http.Response, error) {
				if body == nil {
					body = bytes.NewBuffer([]byte{})
				}

				expectedURL := tt.url(models.Metric{
					Type:  models.Gauge,
					Name:  "Alloc",
					Value: 123,
				})
				assert.Equal(t, expectedURL, url, "Urls should be equal")

				return tt.response, tt.postError
			}

			mc := &MetricsController{
				metricsService: tt.fields.metricsService,
			}

			mc.metricsService.CollectMetrics()

			mc.SendMetrics()

			if tt.postError == nil {
				require.Equal(t, http.StatusOK, tt.response.StatusCode, "Expected status code to be equal")
			} else {
				require.NotNil(t, tt.postError, "Error metric")
			}
		})
	}
}
