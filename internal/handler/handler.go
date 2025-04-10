package handler

import (
	"errors"
	"fmt"
	"github.com/zubans/metrics/internal/errdefs"
	"github.com/zubans/metrics/internal/services"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServerMetricService interface {
	UpdateGauges(mData *services.MetricData) (*errdefs.CustomError, error)
	GetMetrics(mData *services.MetricData) (string, *errdefs.CustomError)
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

	details, err := h.service.UpdateGauges(mData)

	if err != nil {
		var CustomErr *errdefs.CustomError
		if errors.As(details, &CustomErr) {
			http.Error(w, CustomErr.Message, CustomErr.Code)
			fmt.Printf("custom error: %s\n", CustomErr.Error())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Printf("metric updated")
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
	res, err = h.service.GetMetrics(mData)

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

func (h *Handler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	value := h.service.ShowMetrics()

	_, err := io.WriteString(w, value)
	if err != nil {
		return
	}
}
