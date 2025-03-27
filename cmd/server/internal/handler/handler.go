package handler

import (
	"fmt"
	"github.com/zubans/metrics/cmd/server/internal/storage"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage storage.MetricStorage
}

func NewHandler(s storage.MetricStorage) *Handler {
	return &Handler{storage: s}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "metric name required", http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 32)
		if err != nil {
			http.Error(w, "invlid gauge value", http.StatusBadRequest)
			return
		}

		h.storage.UpdateGauge(metricName, value)
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter metric value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, value)
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Printf("metric updated")
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricType == "" || metricName == "" {
		http.Error(w, "metric type, name are required", http.StatusNotFound)
	}

	var res string

	if metricType == "counter" {
		value, found := h.storage.GetCounter(metricName)
		if found {
			res = strconv.FormatInt(value, 10)
		} else {
			http.Error(w, "Counter not found", http.StatusNotFound)
			return
		}
	} else if metricType == "gauge" {
		value, found := h.storage.GetGauge(metricName)
		if found {
			res = strconv.FormatFloat(value, 'f', 1, 64)
		} else {
			http.Error(w, "Gauge not found", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")

	io.WriteString(w, res)
}

func (h *Handler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	value := h.storage.ShowMetrics()

	io.WriteString(w, value)
}
