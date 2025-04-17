package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zubans/metrics/internal/errdefs"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type ServerMetricService interface {
	UpdateMetric(mData *services.MetricData) (*errdefs.CustomError, error)
	GetMetric(mData *services.MetricData) (string, *errdefs.CustomError)
	GetJSONMetric(jsonData *models.MetricsDTO) ([]byte, *errdefs.CustomError)
	ShowMetrics() string
}

type Handler struct {
	service ServerMetricService
}

func NewHandler(service ServerMetricService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "value")
	mData, err := services.NewMetricData(
		chi.URLParam(r, "type"),
		chi.URLParam(r, "name"),
		v,
	)

	if err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	details, err := h.service.UpdateMetric(mData)

	if err != nil {
		var CustomErr *errdefs.CustomError
		if errors.As(details, &CustomErr) {
			http.Error(w, CustomErr.Message, CustomErr.Code)
			logger.Log.Info("custom error",
				zap.String("message", CustomErr.Message),
				zap.Int("status_code", CustomErr.Code),
			)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m models.MetricsDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	mData := &services.MetricData{
		Type: m.MType,
		Name: m.ID,
	}

	switch m.MType {
	case string(models.Gauge):
		val := strconv.FormatFloat(*m.Value, 'f', -1, 64)
		mData.Value = &val
	case string(models.Counter):
		val := strconv.FormatInt(int64(*m.Delta), 10)
		mData.Value = &val
	default:
		http.Error(w, "invalid input", http.StatusBadRequest)
	}

	details, err := h.service.UpdateMetric(mData)

	if err != nil {
		var CustomErr *errdefs.CustomError
		if errors.As(details, &CustomErr) {
			logger.Log.Info("custom error",
				zap.String("message", CustomErr.Message),
				zap.Int("status_code", CustomErr.Code),
			)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	mData, err := services.NewMetricData(
		chi.URLParam(r, "type"),
		chi.URLParam(r, "name"),
	)

	if err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var res string
	res, err = h.service.GetMetric(mData)

	var CustomErr *errdefs.CustomError

	if err.(*errdefs.CustomError) != nil {
		if errors.As(err, &CustomErr) {
			http.Error(w, CustomErr.Message, CustomErr.Code)
			fmt.Printf("custom error: %s\n", CustomErr.Error())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")

	_, err = io.WriteString(w, res)
	if err != nil {
		return
	}
}

func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m *models.MetricsDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var res []byte
	var err error

	res, err = h.service.GetJSONMetric(m)

	var CustomErr *errdefs.CustomError

	if err.(*errdefs.CustomError) != nil {
		if errors.As(err, &CustomErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(CustomErr.Code)

			jsonResp := map[string]interface{}{
				"error":   CustomErr.Message,
				"code":    CustomErr.Code,
				"details": m,
			}
			if err := json.NewEncoder(w).Encode(jsonResp); err != nil {
				logger.Log.Info("failed to encode JSON error response", zap.Error(err))
			}

			logger.Log.Info("custom error",
				zap.String("message", CustomErr.Message),
				zap.Int("status_code", CustomErr.Code),
				zap.Any("body", m),
				zap.Any("raw", err),
				zap.Any("request", r),
			)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(res)
	if err != nil {
		return
	}
}

func (h *Handler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	value := h.service.ShowMetrics()

	_, err := io.WriteString(w, value)
	if err != nil {
		return
	}
}
