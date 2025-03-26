package handler

import (
	"fmt"
	"github.com/zubans/metrics/cmd/server/internal/storage"
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

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	return r
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
		value, err := strconv.ParseFloat(metricValue, 64)
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
