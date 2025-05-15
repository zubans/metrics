package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/zubans/metrics/internal/config"
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
	UpdateMetric(ctx context.Context, mData *services.MetricData) (*models.MetricsDTO, *errdefs.CustomError, error)
	UpdateMetrics(ctx context.Context, m []models.MetricsDTO) (bool, *errdefs.CustomError, error)
	GetMetric(ctx context.Context, mData *services.MetricData) (string, *errdefs.CustomError)
	GetJSONMetric(ctx context.Context, jsonData *models.MetricsDTO) ([]byte, *errdefs.CustomError)
	ShowMetrics(ctx context.Context) (string, error)
	Ping(ctx context.Context) error
}

type Handler struct {
	service ServerMetricService
	cfg     *config.Config
}

func New(service ServerMetricService, cfg *config.Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	_, details, err := h.service.UpdateMetric(ctx, mData)

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

func (h *Handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	var m []models.MetricsDTO
	ctx := r.Context()

	hash := r.Header.Get("HashSHA256")

	if hash != "" {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, "can't read body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))

		h := hmac.New(sha256.New, []byte(h.cfg.Key))
		h.Write(b)
		etolon := h.Sum(nil)

		if hex.EncodeToString(etolon) != hash {
			writeJSONError(w, "incorrect header hash", http.StatusBadRequest)
			return
		}
	}

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSONError(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, details, err := h.service.UpdateMetrics(ctx, m)

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

func (h *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m models.MetricsDTO
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSONError(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	mData := &services.MetricData{
		Type: m.MType,
		Name: m.ID,
	}

	switch m.MType {
	case string(models.Gauge):
		if m.Value == nil {
			writeJSONError(w, "missing gauge value", http.StatusBadRequest)
			return
		}
		if m.Delta != nil {
			writeJSONError(w, "gauge metric should not contain delta", http.StatusBadRequest)
			return
		}

		val := strconv.FormatFloat(*m.Value, 'f', -1, 64)
		mData.Value = &val
	case string(models.Counter):
		if m.Delta == nil {
			writeJSONError(w, "missing counter delta", http.StatusBadRequest)
			return
		}
		if m.Value != nil {
			writeJSONError(w, "counter metric should not contain value", http.StatusBadRequest)
			return
		}

		val := strconv.FormatInt(int64(*m.Delta), 10)
		mData.Value = &val
	default:
		writeJSONError(w, "invalid input", http.StatusBadRequest)
		return
	}

	res, details, err := h.service.UpdateMetric(ctx, mData)

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

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
	}
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	mData, err := services.NewMetricData(
		chi.URLParam(r, "type"),
		chi.URLParam(r, "name"),
	)

	if err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var res string
	res, err = h.service.GetMetric(ctx, mData)

	var CustomErr *errdefs.CustomError

	if err.(*errdefs.CustomError) != nil {
		if errors.As(err, &CustomErr) {
			http.Error(w, CustomErr.Message, CustomErr.Code)
			logger.Log.Info("custom error", zap.String("message", CustomErr.Error()))
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
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var res []byte
	var err error

	res, err = h.service.GetJSONMetric(ctx, m)

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(res)
	if err != nil {
		return
	}
}

func (h *Handler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	value, err := h.service.ShowMetrics(ctx)
	if err != nil {
		logger.Log.Info("failed to get metrics", zap.Error(err))
	}

	_, err = io.WriteString(w, value)
	if err != nil {
		return
	}
}

func (h *Handler) PingServer(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = io.WriteString(w, "")
	if err != nil {
		return
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := map[string]string{
		"error": message,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
